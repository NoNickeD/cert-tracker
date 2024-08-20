package ulfflogger

import (
	"encoding/json"
	"log"
	"time"
)

type ContextInfo map[string]interface{}

func logULFF(level, component, message string, context ContextInfo) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	contextStr, _ := json.Marshal(context)
	log.Printf("%s [%s] [%s] %s Context: %s\n", timestamp, level, component, message, contextStr)
}

func Info(component, message string, context ContextInfo) {
	logULFF("INFO", component, message, context)
}

func Error(component, message string, context ContextInfo) {
	logULFF("ERROR", component, message, context)
}
