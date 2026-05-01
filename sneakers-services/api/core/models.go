package core

type Sneaker struct {
	ID    int
	Brand string
	Model string
}

type CreateSneaker struct {
	Brand string
	Model string
}

type UpdateSneaker struct {
	ID    int64
	Brand string
	Model string
}
