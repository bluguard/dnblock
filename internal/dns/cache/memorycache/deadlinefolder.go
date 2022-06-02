package memorycache

import (
	"log"
	"time"
)

//deadline representation of a deadline
type deadline struct { // estimate cost is 30 bytes
	expiry time.Time
	key    uint32
}

//structure to storee deadlines
type deadlineFolder struct {
	memory []deadline
}

//insert insert a deadline at tyhe right postion
func (f *deadlineFolder) insert(d deadline) {
	f.memory = append(f.memory, d)
	log.Println("deadlines", len(f.memory))
	for i, o := range f.memory {
		if d.expiry.Before(o.expiry) {
			f.shiftRightFrom(i)
			f.memory[i] = d
			return
		}
	}
}

//shiftRightFrom shift to the right from the given position
func (f *deadlineFolder) shiftRightFrom(index int) {
	for i := len(f.memory) - 2; i > index; i-- {
		f.memory[i+1] = f.memory[i]
	}
}

//shiftLeftOf shift to the left of n entries
func (f *deadlineFolder) shiftLeftOf(n int) {
	if n == 0 {
		return // no shift if n == 0
	}
	for i := range f.memory {
		if i < n {
			continue
		}
		f.memory[i-n] = f.memory[i]
	}
	f.memory = f.memory[0 : len(f.memory)-n] // recycle the freed places
}
