package internal

import "context"

type Callback func(context.Context) error
