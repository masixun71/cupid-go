package process

import "strconv"

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

func (f *CompareCheckJob) String() string {

	return "Id:" + strconv.Itoa(int(f.Id)) + ",RetireTime:" + strconv.Itoa(int(f.RetireTime))

}

func (f *CompareCheckJob) Do() error {
	CompareColumn(f.Id)

	return nil
}


