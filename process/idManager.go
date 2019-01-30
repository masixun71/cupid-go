package process

type idManager struct {
	currentId uint
	maxId     uint
}

func (i *idManager) getCurrentId() uint  {
	return i.currentId
}

func (i *idManager) setCurrentId(currentId uint)  {
	i.currentId = currentId
}

func (i *idManager) getMaxId() uint  {
	return i.maxId
}

func (i *idManager) setMaxId(maxId uint)  {
	i.maxId = maxId
}

func (i *idManager) incrCurrentId() {
	i.currentId++
}

func (i *idManager) hasNewId() bool {
	return i.currentId <= i.maxId
}

