package storage

type ErrLoginExist struct {
}

func (e *ErrLoginExist) Error() string { return "login already exist" }

type ErrAlreadyLoadedByThisUser struct {
}

func (e *ErrAlreadyLoadedByThisUser) Error() string {
	return "order has been loaded already by this user"
}

type ErrAlreadyLoadedByDifferentUser struct {
}

func (e *ErrAlreadyLoadedByDifferentUser) Error() string {
	return "order has been loaded already by different user"
}

type ErrDBInteraction struct {
}

func (e *ErrDBInteraction) Error() string { return "database interaction error" }

type ErrFormat struct {
}

func (e *ErrFormat) Error() string { return "wrong value format" }
