package config

const LogFile = "LogFile"
const IsDev = "IsDev"
const DBConnString = "DBConnString"
const Port = "Port"

var defaults = map[string]string{
	LogFile:      "logs.txt",
	IsDev:        "true",
	DBConnString: "",
	Port:         "8000",
}
