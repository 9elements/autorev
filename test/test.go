package test

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/9elements/autorev/tracelog"

	"github.com/9elements/autorev/config"
)

// test - Struct which holds relevant information for test
type test struct {
	// Latest Test id we fetched
	LatestTestID int
	// Database Connection
	db *sql.DB
	// Config
	cfg config.Config
}

// Close - Close DB Connection
func (t *test) Close() {
	t.db.Close()
}

// Init = Initialize a new Test
func Init(cfg config.Config) (*test, error) {
	var err error

	t := test{
		LatestTestID: -1,
		db:           nil,
	}
	t.cfg = cfg

	log.Println("Seting up database connection..")
	// Setup Mysql Connection
	t.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/autorev", t.cfg.Database.Username, t.cfg.Database.Password))

	if err != nil {
		return nil, err
	}
	err = t.db.Ping()
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// GetLatestTestID - Fetch Latest Test ID from Struct
func (t *test) GetLatestTestID() int {
	return t.LatestTestID
}

// SetLatestTestID - Set Latest Test ID in Struct
func (t *test) SetLatestTestID(val int) {
	t.LatestTestID = val
}

// GetNextTest - Get next free test
func (t *test) GetNextTest() (int, error) {
	if t.db == nil {
		return -1, fmt.Errorf("DB Function Pointer is nil")
	}

	// Fetch latest test, update and grab upd config
	nextTest, err := t.db.Query("SELECT idTests FROM tests WHERE status = 0 ORDER BY ts_added ASC LIMIT 1;")
	if err != nil {
		return -1, err
	}

	columns, err := nextTest.Columns()

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	if nextTest.Next() {
		err = nextTest.Scan(scanArgs...)
	} else {
		return 0, nil
	}

	t.LatestTestID, err = strconv.Atoi(string(values[0]))
	if err != nil {
		return -1, err
	}
	log.Printf("Next Test: %d\n", t.LatestTestID)

	return t.LatestTestID, nil
}

// SetTestInProgress - Update Test to be in Progress
func (t *test) SetTestInProgress() error {

	stmtUpdate, err := t.db.Prepare("UPDATE tests SET status = 1, ts_started = NOW() WHERE idTests = ?")
	if err != nil {
		return err
	}
	stmtUpdate.Exec(t.LatestTestID)

	return nil
}

// SetTestSuccessful - Update Test to failed
func (t *test) SetTestSuccessful() error {

	stmtUpdate, err := t.db.Prepare("UPDATE tests SET status = 2, ts_finished = NOW() WHERE idTests = ?")
	if err != nil {
		return err
	}
	stmtUpdate.Exec(t.LatestTestID)

	return nil
}

// SetTestFailed - Update Test to failed
func (t *test) SetTestFailed() error {

	stmtUpdate, err := t.db.Prepare("UPDATE tests SET status = 3, ts_finished = NOW() WHERE idTests = ?")
	if err != nil {
		return err
	}
	stmtUpdate.Exec(t.LatestTestID)

	return nil
}

// GenNewTest - Insert a test into DB
func (t *test) GenNewTest(name string, config config.Config, configBlob []byte) error {
	if t.db == nil {
		return fmt.Errorf("DB Function Pointer is nil")
	}

	stmt, err := t.db.Prepare("INSERT INTO tests (status, ts_added, config, fk_defaultConfig) VALUES (0, NOW(), ?, (SELECT updId FROM updDefaults WHERE platformName = ?))")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(configBlob, config.TraceLog.OptionsDefaultTable)
	if err != nil {
		return err
	}
	return nil
}

// SetNewDefaultConfig - Add a new default config with name to the DB
func (t *test) SetNewDefaultConfig(name string, config []byte) error {
	/*
		if t.db == nil {
			return fmt.Errorf("DB Function Pointer is nil")
		}

		if t.LatestTestID == -1 {
			return fmt.Errorf("Latest Test ID is not set")
		}
	*/

	// TODO: Check if platform already exists - ask to overwrite, cancel or new name

	stmt, err := t.db.Prepare("INSERT INTO `updDefaults` (`platformName`, `size`, `configBlob`) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(name, len(config), config)
	if err != nil {
		return err
	}
	log.Printf("Inserted new config as platform %s\n", name)

	return nil
}

// GetDefaultConfig - Fetches the default config for a given name
func (t *test) GetDefaultConfig(name string) ([]byte, error) {
	if t.db == nil {
		return nil, fmt.Errorf("DB Function Pointer is nil")
	}

	// Get Size of DefaultConfig
	stmt, err := t.db.Prepare("SELECT size FROM updDefaults WHERE platformName = ?")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var amount int

	err = stmt.QueryRow(name).Scan(&amount)
	if err != nil {
		return nil, err
	}

	// Fetch Default Config
	config := make([]byte, amount)

	scanArgs := make([]interface{}, 1)
	scanArgs[0] = &config[0]

	stmtConfig, err := t.db.Prepare("SELECT configBlob FROM updDefaults WHERE platformName = ? LIMIT 1")

	if err != nil {
		return nil, err
	}

	defer stmtConfig.Close()

	err = stmtConfig.QueryRow(name).Scan(&config)
	if err != nil {
		return nil, err
	}

	return config, nil

}

// GetConfig - Fetch the config from the last test
func (t *test) GetConfig(testID int) ([]byte, error) {
	if t.db == nil {
		return nil, fmt.Errorf("DB Function Pointer is nil")
	}

	if testID == -1 {
		return nil, fmt.Errorf("Invalid testID")
	}

	// Get Size of DefaultConfig
	stmt, err := t.db.Prepare("SELECT size FROM updDefaults WHERE platformName = ?")

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var amount int
	name := t.cfg.TraceLog.OptionsDefaultTable

	err = stmt.QueryRow(name).Scan(&amount)
	if err != nil {
		return nil, err
	}

	// Fetch Default Config
	config := make([]byte, amount)

	scanArgs := make([]interface{}, 1)
	scanArgs[0] = &config[0]

	stmtConfig, err := t.db.Prepare("SELECT config FROM tests WHERE idTests = ? LIMIT 1")

	if err != nil {
		return nil, err
	}

	defer stmtConfig.Close()

	err = stmtConfig.QueryRow(testID).Scan(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// WriteSetIntoDB - write a TraceLogEntry Set into the DB with fk = LasttestTestID123
func (t *test) WriteSetIntoDB(entries []tracelog.TraceLogEntry) error {

	stmt, err := t.db.Prepare("INSERT INTO traceLog (type, input, address, value, ip, accessSize, fk_idTests) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err = stmt.Exec(entry.Type, entry.Inout, entry.Address, entry.Value, entry.IP, entry.AccessSize, t.LatestTestID)
		if err != nil {
			return err
		}
	}

	return nil
}

// FetchTraceLogEntriesFromDB - Fetches TraceLogEntries from the DB for a given test testID
func (t *test) FetchTraceLogEntriesFromDB(testID int) ([]tracelog.TraceLogEntry, error) {
	if t.db == nil {
		return nil, fmt.Errorf("DB Function Pointer is nil")
	}

	var traceLogEntries []tracelog.TraceLogEntry

	rows, err := t.db.Query("SELECT type, input, address, value, ip, accessSize FROM traceLog WHERE fk_idTests = ? ORDER BY idTraceLog ASC", testID)
	if err != nil {
		return nil, err
	}
	values := make([]uint64, 6)
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for {
		if rows.Next() {
			err = rows.Scan(scanArgs...)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}

		tracelogentry := tracelog.TraceLogEntry{
			Type:       int(values[0]),
			Inout:      (values[1] != 0),
			Address:    uint(values[2]),
			Value:      uint64(values[3]),
			IP:         uint(values[4]),
			AccessSize: uint(values[5]),
		}
		traceLogEntries = append(traceLogEntries, tracelogentry)
	}
	return traceLogEntries, nil
}

// FetchTraceLogEntriesFromDB - Fetches TraceLogEntries from the DB for a given test testID
func (t *test) FetchSuccessfulTraceLogIDFromDB() ([]int, error) {
	if t.db == nil {
		return nil, fmt.Errorf("DB Function Pointer is nil")
	}

	var traceLogIDs []int

	rows, err := t.db.Query("SELECT idTests FROM tests WHERE status = 2")
	if err != nil {
		return nil, err
	}
	values := make([]uint64, 1)
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for {
		if rows.Next() {
			err = rows.Scan(scanArgs...)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}

		traceLogIDs = append(traceLogIDs, int(values[0]))
	}
	return traceLogIDs, nil
}

// GenRecursiveNewTestsFromCfg - Recursivly creates new tests based on user provided FirmwareOption config
func (t *test) GenRecursiveNewTestsFromCfg(name string, cfg config.Config, blob []byte, level int) (uint, error) {
	var cnt uint
	if level > len(cfg.TraceLog.VariableFirmareOptions) {
		// should not happen
		return cnt, nil
	} else if level == len(cfg.TraceLog.VariableFirmareOptions) {
		err := t.GenNewTest(name, cfg, blob)
		cnt++
		return cnt, err
	}

	opt := cfg.TraceLog.VariableFirmareOptions[level]

	if opt.ByteOffset > uint(len(blob)) {
		log.Printf("Invalid ByteOffset specified for %s\n", opt.Name)
	}
	if opt.BitWidth > 64 {
		log.Printf("Invalid BitWidth specified for %s\n", opt.Name)
	}
	if opt.Max < opt.Min {
		log.Printf("Invalid BitWidth specified for %s\n", opt.Name)
	}

	for j := opt.Min; j <= opt.Max; j++ {
		blobcopy := blob

		for a := uint(0); a*8 < opt.BitWidth; a++ {
			blobcopy[opt.ByteOffset+a] = byte((uint64(j) >> (a * 8)))
		}

		i, err := t.GenRecursiveNewTestsFromCfg(name+fmt.Sprintf("%s=%d ", opt.Name, j), cfg, blobcopy, level+1)
		cnt += i
		if err != nil {
			return cnt, err
		}
	}

	return cnt, nil
}

// GetFirmwareOptionsFromConfigBLOBs - Convert blob config of test "testID" to map of UPDs
func (t *test) GetFirmwareOptionsFromConfigBLOBs(cfg config.Config, testID int) (map[string]uint64, error) {
	optionsset := map[string]uint64{}

	currentblob, err := t.GetConfig(testID)
	if err != nil {
		return nil, err
	}

	for _, opt := range cfg.TraceLog.VariableFirmareOptions {
		if opt.ByteOffset > uint(len(currentblob)) {
			log.Printf("Invalid ByteOffset specified for %s\n", opt.Name)
			continue
		}
		if opt.BitWidth > 64 {
			log.Printf("Invalid BitWidth specified for %s\n", opt.Name)
			continue
		}
		if opt.Max < opt.Min {
			log.Printf("Invalid BitWidth specified for %s\n", opt.Name)
			continue
		}

		var val uint64
		for a := uint(0); a*8 < opt.BitWidth; a++ {
			val |= uint64(currentblob[opt.ByteOffset+a]) << (a * 8)
		}
		optionsset[opt.Name] = val
	}

	return optionsset, nil
}
