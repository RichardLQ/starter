package ginv

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	_ "github.com/rs/zerolog/log"
	"net/http"
	"reflect"
)


// DefaultHandlerWrapper 全局Wrapper配置
var DefaultHandlerWrapper = HandlerWrapper{}


// HandlerWrapper wrapper定义
type HandlerWrapper struct {
	RspRenderer func(c *gin.Context, rsp interface{}, err error)
}

var typeContext = reflect.TypeOf(new(context.Context)).Elem()
var typeError = reflect.TypeOf(new(error)).Elem()

func WrapHandler(h interface{}) gin.HandlerFunc {
	if h == nil {
		panic("WrapHandler不能传入空")
	}
	t := reflect.TypeOf(h)
	if t.NumIn() < 3 {
		panic("入参必须填三个参数")
	}
	if reflect.TypeOf(h).Kind() != reflect.Func {
		panic("入参必须为函数")
	}
	if typeContext.AssignableTo(t.In(0)) && t.Out(0).AssignableTo(typeError) {
		return wrapType31(reflect.ValueOf(h))
	}
	return panic("未找到此函数")
}

func wrapType31(hv reflect.Value) gin.HandlerFunc {
	t := hv.Type()
	return func(c *gin.Context) {
		var reqValue []reflect.Value
		var rspV reflect.Value
		reqValue = append(reqValue, reflect.ValueOf(c.Request.Context()))
		for i := 1; i < t.NumIn(); i++ {
			reqV, _ := inparam(c, t.In(i).Elem())
			reqValue = append(reqValue, reqV)
			if i == 2 {
				rspV = reqV
			}
		}
		out := hv.Call(reqValue)
		DefaultRspRenderer(c,rspV.Interface(),out[0].Interface().(error))
	}
}

func inparam(c *gin.Context, reqs reflect.Type) (reflect.Value, error) {
	var bs []binding.Binding
	b := binding.Default(c.Request.Method, c.ContentType())
	if b == binding.Form {
		bs = []binding.Binding{b, binding.Header}
	} else {
		bs = []binding.Binding{b, binding.Header, binding.Form}
	}
	reqV := reflect.New(reqs)
	req := reqV.Interface()
	for _, b := range bs {
		err := c.ShouldBindWith(req, b)
		if err != nil {
			return reqV, err
		}
	}
	err := c.ShouldBindUri(req)
	if err != nil {
		return reflect.Value{}, err
	}
	return reqV, nil
}


// DefaultRspRenderer 默认的请求结果处理函数
func DefaultRspRenderer(c *gin.Context, rsp interface{}, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, rsp)
}