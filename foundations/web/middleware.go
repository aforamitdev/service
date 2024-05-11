package web

import "fmt"

type Middleware func(Handler) Handler

func wrapMiddleware(mw []Middleware, handler Handler, marker string) Handler {

	fmt.Println(len(mw), marker)

	for i := len(mw) - 1; i >= 0; i-- {

		h := mw[i]

		if h != nil {
			handler = h(handler)
		}

	}

	return handler

}
