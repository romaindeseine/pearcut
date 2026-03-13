package main

type ReadStore interface {
	Get(slug string) (Experiment, error)
	List(filter ExperimentFilter) ([]Experiment, error)
}

type WriteStore interface {
	Create(exp Experiment) error
	Update(exp Experiment) error
	Delete(slug string) error
}

type Store interface {
	ReadStore
	WriteStore
}
