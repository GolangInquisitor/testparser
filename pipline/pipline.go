// pipline project pipline.go
package pipline

import (
	"github.com/gocolly/colly"
	"context"
)

type (
	HTMLElemChan  chan *colly.HTMLElement
	
	HTMLElemPline struct {
	pipline 	chan *colly.HTMLElement
	handler HTMLCallback
	
	}
)
func (p *HTMLElemPline) Put(e *colly.HTMLElement){
	
}
func NewPipline(ctx colly.Request.Ctx, len int, pipehandler HTMLCallback) (HTMLElemPline, error) {
  
}
