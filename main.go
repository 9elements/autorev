package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/9elements/autorev/config"
	"github.com/9elements/autorev/ir"
	"github.com/9elements/autorev/mesh"
	"github.com/9elements/autorev/test"
	"github.com/9elements/autorev/tracelog"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	devTTYDevicePath := flag.String("dev", "", "The serial device to use for communication with autorev shell")
	baudTTYDevice := flag.Int("baud", 115200, "The serial device baud rate to use")
	fifoDevicePath := flag.String("fifo", "", "The fifos to communicate with a debug target (appends .in and .out)")
	collectNewTrace := flag.Bool("runtrace", false, "Collect a new tracelog")
	addNewTrace := flag.Bool("newtrace", false, "Add new tracelogs based on FirmwareOption config")
	addNewConfig := flag.Bool("newConfig", false, "Add new default config")
	newConfigName := flag.String("newConfigName", "", "Name of the new default config")
	newConfigFile := flag.String("newConfigFile", "", "Path to nee default config file")
	collectAllTraces := flag.Bool("collecttraces", false, "Collects all tracelogs that haven't run yet")
	buildAst := flag.Bool("buildast", false, "Generates an AST from all successful tracelogs")
	genCCode := flag.String("genCcode", "", "Path to generated C code from AST. To be used with -buildAst")
	genDot := flag.String("genDot", "", "Path to generated dot file from AST. To be used with -buildAst")

	verbose := flag.Bool("verbose", false, "Be verbose")

	flag.Parse()

	cfg, err := config.GetConfig()

	if err != nil {
		panic(err.Error())
	}

	if len(cfg.Database.Password) == 0 {
		log.Printf("Database password is empty in config.yml!")
		log.Printf("Did you set up a database already?")
	}

	test, err := test.Init(cfg)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	defer test.Close()

	if *addNewConfig { // Add a new Config aka FirmwareOption BLOB to DB

		// Open config file
		f, err := ioutil.ReadFile(*newConfigFile)
		if err != nil {
			log.Printf("%v\n", err)
			return
		}

		err = test.SetNewDefaultConfig(*newConfigName, f)
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
	} else if *collectAllTraces { // Process all tracelogs not run yet

		tl, err := tracelog.CreateTraceLog(*devTTYDevicePath, *baudTTYDevice, *fifoDevicePath, cfg)
		if err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		}
		tl.SetVerbose(*verbose)
		// The complete mesh

		for {
			id, err := test.GetNextTest()
			if err != nil {
				log.Printf("%v\n", err)
				return
			}
			if id == 0 {
				log.Println("All test have been run.")
				break
			}
			config, err := test.GetConfig(test.LatestTestID)
			if err != nil {
				log.Printf("%v\n", err)
				return
			}
			err = test.SetTestInProgress()
			if err != nil {
				log.Printf("%v\n", err)
				return
			}

			tles, err := tl.CollectNewTracelog(config)
			if err != nil {
				log.Printf("%v\n", err)
				err = test.SetTestFailed()
				if err != nil {
					log.Printf("%v\n", err)
				}
				continue
			}
			err = test.SetTestSuccessful()
			if err != nil {
				log.Printf("%v\n", err)
			}
			log.Printf("Writing to DB..")
			err = test.WriteSetIntoDB(tles)
			if err != nil {
				log.Printf("%v", err)
				return
			}
			log.Println("Done.")
		}

	} else if *collectNewTrace { // Collect a single new tracelog that haven't run yet

		tl, err := tracelog.CreateTraceLog(*devTTYDevicePath, *baudTTYDevice, *fifoDevicePath, cfg)
		if err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		}
		tl.SetVerbose(*verbose)

		_, err = test.GetNextTest()
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
		config, err := test.GetConfig(test.LatestTestID)
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
		err = test.SetTestInProgress()
		if err != nil {
			log.Printf("%v\n", err)
			return
		}
		tles, err := tl.CollectNewTracelog(config)
		if err != nil {
			log.Printf("%v\n", err)
			err = test.SetTestFailed()
			if err != nil {
				log.Printf("%v\n", err)
			}
			os.Exit(1)
		}
		err = test.SetTestSuccessful()
		if err != nil {
			log.Printf("%v\n", err)
		}
		log.Printf("Writing %d lines into the DB..", len(tles))
		err = test.WriteSetIntoDB(tles)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		log.Println("Done.")
		f, err := os.Create("generatedC.c")
		if err != nil {
			log.Printf("%v", err)
			return
		}

		for i := range tles {
			if tles[i].Inout {
				p := ir.IRNewPrimitiveRead(tles[i])
				line := fmt.Sprintf("%s", p.ConvertToC())
				f.WriteString(line)
				fmt.Println(line)
			} else {
				p := ir.IRNewPrimitiveWrite(tles[i])
				line := fmt.Sprintf("%s", p.ConvertToC())
				f.WriteString(line)
				fmt.Println(line)
			}
		}

	} else if *addNewTrace { // Generate recursive tracelogs to be run based on config.yml

		blob, err := test.GetDefaultConfig(cfg.TraceLog.OptionsDefaultTable)
		if err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		}

		cnt, err := test.GenRecursiveNewTestsFromCfg("", cfg, blob, 0)
		if err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		}
		log.Printf("Added %d new tracelogs to be tested\n", cnt)
	} else if *buildAst {
		testIds, err := test.FetchSuccessfulTraceLogIDFromDB()
		if err != nil {
			log.Printf(err.Error())
			os.Exit(1)
		}
		if len(testIds) == 0 {
			log.Printf("No tests in database, aborting\n")
			os.Exit(1)
		}
		var m = mesh.Mesh{Start: mesh.MeshNode{Id: 0, Hash: "0"}}

		for t := range testIds {
			log.Printf("Merging test id %d\n", testIds[t])
			options, err := test.GetFirmwareOptionsFromConfigBLOBs(cfg, testIds[t])
			if err != nil {
				log.Printf(err.Error())
				os.Exit(1)
			}
			log.Printf("%v\n", options)
			tles, err := test.FetchTraceLogEntriesFromDB(testIds[t])
			if err != nil {
				log.Printf(err.Error())
				os.Exit(1)
			}
			log.Printf(" %d trace log entries\n", len(tles))

			err = m.InsertTraceLogIntoMesh(tles, options)
			if err != nil {
				log.Printf(err.Error())
				os.Exit(1)
			}
		}
		allFirmwareOptions := config.GetConfigFirmwareOptionsByName(cfg)
		m.OptimiseMeshByRemovingNodes()
		m.OptimiseMeshByRemovingFirmwareOptions(allFirmwareOptions)
		// Experimental mesh optimisation...
		m.OptimiseMeshByAddingNops()
		if len(*genCCode) > 0 {
			err := ioutil.WriteFile(*genCCode, []byte(ir.MeshToIR(&m)), 0644)
			if err != nil {
				log.Printf(err.Error())
				os.Exit(1)
			}
		}
		if len(*genDot) > 0 {
			m.WriteDot(*genDot)
		}
	} else {
		log.Println("Error: No action given! Nothing to do.")
		flag.Usage()
	}
}
