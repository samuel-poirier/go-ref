package app

type Consumer interface {
  StartConsuming() error
}
