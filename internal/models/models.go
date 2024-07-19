package models

type Request struct {
	Parallel bool `json:"parallel"`
}

type Response struct {
	Completion string `json:"completion"`
}
