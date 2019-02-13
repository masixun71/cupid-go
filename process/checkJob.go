package process

type CompareCheckJob struct {
	Id         uint
	RetireTime int
}

func (f *CompareCheckJob) GetRetireTime() int {
	return f.RetireTime
}

func (f *CompareCheckJob) IncrRetireTime() {
	f.RetireTime++
}


func (f *CompareCheckJob) Do() error {
	CompareColumn(f.Id)

	return nil
}


