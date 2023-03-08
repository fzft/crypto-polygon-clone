package core

import "io"

type Encoder[T any] interface {
	Encode(writer io.Writer, value T) error
}

type Decoder[T any] interface {
	Decode(reader io.Reader, value T) error
}


