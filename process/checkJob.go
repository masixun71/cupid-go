package process

type CompareCheckJob struct {
	id uint
}

func (f *CompareCheckJob) Do() error {
	CompareColumn(f.id)

	return nil
}


