package errors

type MissingTestAgentError struct {
   Err error
}

func (e *MissingTestAgentError) Error() string {
   return "Missing or invalid TestAgent"
}

type MissingMTAgentError struct {
   Err error
}

func (e *MissingMTAgentError) Error() string {
   return "Missing or invalid MTAgent"
}

type NoValidTestsError struct {
   Err error
}

func (e *NoValidTestsError) Error() string {
   return "No valid Tests provided, nothing to do"
}

type MessageNotFound struct {
   Msg string
}

func (e *MessageNotFound) Error() string {
   return e.Msg
}
