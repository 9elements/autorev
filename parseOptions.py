import argparse
# Read in UPD File and generate DB Table for it.

types = {
        "UINT8"     : "TINYINT",
        "UINT16"    : "SMALLINT",
        "UINT32"    : "INT",
        "UINT64"    : "BIGINT"
        }

readBytes = {
        "UINT8"     : 1,
        "UINT16"    : 2,
        "UINT32"    : 4,
        "UINT64"    : 8
        }

parser = argparse.ArgumentParser(description='Parse UPD Options into database \
        code.')
parser.add_argument('platform', type=str, help='Name of the Platform')
args = parser.parse_args()

updFile = open("updDescKabylake.txt", "r")
dbFile = open("updDefaults_" + args.platform + ".sql", "w")
defaultFile = open("fspsKabyLakeDefault.bin", "rb")

dbFile.write("CREATE TABLE IF NOT EXISTS `autorev`.`updDefaults_" +
        args.platform + "` (\n")
dbFile.write("\t `" + args.platform + "Id` INT UNSIGNED NOT NULL \
AUTO_INCREMENT,\n")
dbFile.write("\t `name` TEXT NOT NULL,\n")
dbFile.write("\t `offset` VARCHAR(6) NOT NULL DEFAULT '0x0000',\n")
dbFile.write("\t `defaultValue` BIGINT UNSIGNED NOT NULL DEFAULT '0',\n")
dbFile.write("\t `valueSize` INT UNSIGNED NOT NULL DEFAULT '1',\n")
dbFile.write("\t PRIMARY KEY (`" + args.platform + "Id`)\n")
dbFile.write(") ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 \
        COLLATE=utf8mb4_0900_ai_ci;\n")
dbFile.write("\n--\n")
dbFile.write("-- Dumping data for table `updDefaults_" + args.platform + "`\n")
dbFile.write("--\n\n")
dbFile.write("LOCK TABLES `updDefaults_" + args.platform + "` WRITE;\n")
dbFile.write("/*!40000 ALTER TABLE `updDefaults_" + args.platform + "` DISABLE KEYS */;\n")
dbFile.write("INSERT INTO `updDefaults_" + args.platform + "` (`name`, \
        `offset`, `defaultValue`, `valueSize`) VALUES ")

lastOffset = 0
OffsetFound = 0
first = 1

for line in updFile:
    defaultFile.seek(0)
    #Find a new Offset
    if line.strip().startswith("/** Offset"):
        if OffsetFound:
            print("Found Offset two times - Trashing current Offset, Using new"
                        + " one as Offset.")
        lastOffset = line[line.find("0x"):line.find("0x")+6]
        OffsetFound = 1
    if line.strip().startswith("UINT"):
        OffsetFound = 0
        parts = line.strip().split(" ")
        # Part 0 is always filed type - part -1 is always name
        if ("Unused" in parts[-1]) or ("Reserved" in parts[-1]):
            continue
        if first == 0:
            dbFile.write(","),
        first = 0
        if "[" in parts[-1]:
            arraySize = parts[-1][parts[-1].find("[") + 1:parts[-1].find("]")]
            updOption = parts[-1][:parts[-1].find("[")]
            for i in range(int(arraySize)):
                # Read Default Value from Binary
                offset = int(lastOffset, 16)
                offsetArray = offset + (i * readBytes[parts[0]])
                defaultFile.seek(offsetArray)
                offsetArray = "{0:#0{1}x}".format(offsetArray,6)
                val = int.from_bytes(defaultFile.read(readBytes[parts[0]]),
                        byteorder='big')
                defaultFile.seek(0)

                # Write DB Scheme File
                dbFile.write("('%s_%i', '%s', %i, %i)" % (updOption, i, str(offsetArray),
                    val, readBytes[parts[0]]))
                if i != (int(arraySize) -1):
                    dbFile.write(","),
        else:
            offset = int(lastOffset, 16)
            defaultFile.seek(offset)
            val = int.from_bytes(defaultFile.read(readBytes[parts[0]]),
                    byteorder='big')
            offset = "{0:#0{1}x}".format(offset,6)

            parts[-1] = parts[-1].replace(";", "")
            dbFile.write("('%s', '%s', %i, %i)" % (parts[-1], str(offset),
                    val, readBytes[parts[0]]))


dbFile.write(";\n\n")
dbFile.write("UNLOCK TABLES;")
