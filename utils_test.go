package agency

import (
	"fmt"
	"testing"
)

func Test_Tasks(b *testing.T) {
	ts := taskSlice([]*task{})
	ts.push(&task{
		index: 1,
	})
	fmt.Println(ts.pop(), ts.length())
	ts.push(&task{
		index: 2,
	})
	ts.push(&task{
		index: 3,
	})
	ts.push(&task{
		index: 4,
	})
	fmt.Println(ts.remove(1), ts.length(), ts.list())
	for _, t := range ts.list() {
		fmt.Println(t.index)
	}
}
