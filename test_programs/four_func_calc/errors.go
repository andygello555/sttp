package main

type RuntimeError string

func (r RuntimeError) Error() string {
	return string(r)
}

