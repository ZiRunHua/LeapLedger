package globalTask

import "context"

type transactionHandle[Data any] func(Data, context.Context) error

type handle[Data any] func(Data) error
