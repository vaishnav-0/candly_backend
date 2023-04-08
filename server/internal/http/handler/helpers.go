package handler


func AppendJSONMessage(message string, extra map[string]interface{}) map[string]interface{} {
	extra["message"] = message
	return extra
}

func JSONMessage(message string) map[string]interface{} {
	return map[string]interface{}{
		"message":message,
	}

}