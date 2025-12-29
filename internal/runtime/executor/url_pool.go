package executor

import (
	"strings"
	"sync"
)

type URLBuilder struct {
	sb strings.Builder
}

var urlBuilderPool = sync.Pool{
	New: func() any {
		return &URLBuilder{
			sb: strings.Builder{},
		}
	},
}

func GetURLBuilder() *URLBuilder {
	return urlBuilderPool.Get().(*URLBuilder)
}

func (b *URLBuilder) Release() {
	b.sb.Reset()
	urlBuilderPool.Put(b)
}

func (b *URLBuilder) WriteString(s string) {
	b.sb.WriteString(s)
}

func (b *URLBuilder) String() string {
	return b.sb.String()
}

func (b *URLBuilder) Grow(n int) {
	b.sb.Grow(n)
}
