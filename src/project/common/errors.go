package common

const (
	ErrUnknown = iota
	ErrNotFound
	ErrNotImplemented
	ErrReqInvalid
	ErrDatabase
	ErrNotAuth
	ErrAuthFailed
	ErrBanned
	ErrItemNotFound
	ErrAlredyAuth
	ErrEmailInvalid
	ErrEmailRegistered
	ErrCodeSentRecently
	ErrCodeInvalid
	ErrEmailNotFound
	ErrIncorrectPassword
)

var Error = [...]ErrorResponse{
	{false, Err{"unknown_error", "Unknown error"}},
	{false, Err{"not_found", "Method not found"}},
	{false, Err{"not_implemented", "Method not implemented"}},
	{false, Err{"req_invalid", "The request is not valid"}},
	{false, Err{"database_error", "Database Error"}},
	{false, Err{"not_auth", "Auth token required"}},
	{false, Err{"auth_failed", "Invalid token"}},
	{false, Err{"banned", "You are banned"}},
	{false, Err{"item_not_found", "Item not found"}},
	{false, Err{"alredy_auth", "Current user have other account"}},
	{false, Err{"email_invalid", "Incorrect email"}},
	{false, Err{"email_registered", "That email is alredy registered"}},
	{false, Err{"code_sent_recently", "Code was sent recently"}},
	{false, Err{"code_invalid", "Incorrect confirmation code"}},
	{false, Err{"email_not_found", "Not found user with that email"}},
	{false, Err{"incorrect_password", "Incorrect password"}},
}
