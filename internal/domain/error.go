package domain

const (
	UserNotFound             = "USER_NOT_FOUND"
	UserAlreadyExists        = "USER_ALREADY_EXISTS"
	PlanRequestAlreadyExists = "PLAN_REQUEST_ALREADY_EXISTS"
	PlanRequestNotFound      = "PLAN_REQUEST_NOT_FOUND"
	VacanciesNotFound        = "VACANCIES_NOT_FOUND"
	VacancyNotFound          = "VACANCY_NOT_FOUND"
	VacanciesAlreadyExist    = "VACANCIES_ALREADY_EXIST"
	VacanciesAlreadyRespond  = "VACANCIES_ALREADY_RESPOND"
	InvalidCredentials       = "INVALID_CREDENTIALS"
	RedisNotFound            = "REDIS_NOT_FOUND"
	DirectionNotFound        = "DIRECTION_NOT_FOUND"
	RespondAlreadyExists     = "RESPOND_ALREADY_EXISTS"
	VacancyAlreadyExists     = "VACANCY_ALREADY_EXISTS"
	RespondVacanciesNotFound = "RESPOND_VACANCIES_NOT_FOUND"
	RespondVacancyNotFound   = "RESPOND_VACANCY_NOT_FOUND"
	FileTooLarge             = "FILE_TOO_LARGE"
	InvalidContentType       = "INVALID_CONTENT_TYPE"
	DirectionHasVacancies    = "DIRECTION_HAS_VACANCIES"
	TokenExpired             = "TOKEN_EXPIRED"
	InvalidToken             = "INVALID_TOKEN"
	TokenRevoked             = "TOKEN_REVOKED"
	InvalidParam             = "INVALID_PARAM"
)

type DomainError struct {
	code string
	msg  string
}

func (e *DomainError) Error() string {
	return e.msg
}

func (e *DomainError) Code() string {
	return e.code
}

func newError(code, msg string) error {
	return &DomainError{code: code, msg: msg}
}

var (
	ErrUserNotExists             = newError(UserNotFound, "user not exists")
	ErrUserAlreadyExists         = newError(UserAlreadyExists, "user already exists")
	ErrPlanRequestAlreadyExists  = newError(PlanRequestAlreadyExists, "plan request already exists")
	ErrPlanRequestNotExists      = newError(PlanRequestNotFound, "plan request does not exists")
	ErrVacanciesNotExists        = newError(VacanciesNotFound, "vacancies not exists")
	ErrVacancyNotExists          = newError(VacancyNotFound, "vacancy not exists")
	ErrVacanciesAlreadyExists    = newError(VacanciesAlreadyExist, "vacancies already exists")
	ErrVacanciesAlreadyRespond   = newError(VacanciesAlreadyRespond, "vacancies already respond")
	ErrInvalidCredentials        = newError(InvalidCredentials, "invalid credentials")
	ErrRedisNotFound             = newError(RedisNotFound, "redis not found")
	ErrDirectionNotFound         = newError(DirectionNotFound, "direction not found")
	ErrRespondAlreadyExists      = newError(RespondAlreadyExists, "respond already exists")
	ErrVacancyAlreadyExists      = newError(VacancyAlreadyExists, "vacancy already exists")
	ErrRespondVacanciesNotExists = newError(RespondVacanciesNotFound, "respond vacancies does not exists")
	ErrRespondVacancyNotExists   = newError(RespondVacancyNotFound, "respond vacancy does not exists")
	ErrFileTooLarge              = newError(FileTooLarge, "file too large")
	ErrInvalidContentType        = newError(InvalidContentType, "invalid content type")
	ErrDirectionHasVacancies     = newError(DirectionHasVacancies, "conflict with already created vacancies")
	ErrExpiredToken              = newError(TokenExpired, "expired token")
	ErrInvalidToken              = newError(InvalidToken, "invalid token")
	ErrTokenRevoked              = newError(TokenRevoked, "token revoked")
	ErrInvalidParam              = newError(InvalidParam, "invalid param")
)
