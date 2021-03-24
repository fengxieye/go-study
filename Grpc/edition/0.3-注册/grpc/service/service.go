package service

import (
	"go/ast"
	//"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

/*
the method’s type is exported. – 方法所属类型是导出的。
the method is exported. – 方式是导出的。
the method has two arguments, both exported (or builtin) types. – 两个入参，均为导出或内置类型。
the method’s second argument is a pointer. – 第二个入参必须是一个指针。
the method has return type error. – 返回值为 error 类型。
*/
func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		//指针的元素type
		argv = reflect.New(m.ArgType.Elem())
	} else {
		//指针对象的值
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	//返回值为指针
	replyv := reflect.New(m.ReplyType.Elem())
	//log.Println("new Reply type:", m.ReplyType, ",elem kind", m.ReplyType.Elem().Kind(), ",elem type", m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	//eg:m.ReplyType*map[string]int, elem():map[string]int, Elem().Kind():map
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}

	return replyv
}

type service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value //receiver
	method map[string]*methodType
}

func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	//whether name starts with an upper-case letter.
	if !ast.IsExported(s.name) {
		log.Fatalf("newservice: %s is not valid service name", s.name)
	}
	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	//遍历结构体可导出的方法 NumMethod returns the number of exported methods in the type's method set.
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		log.Printf("service.registermethod :check %s.%s\n", s.name, method.Name)
		//方法类型
		mType := method.Type
		//参数和返回值
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		//*error 的元素类型 error,todo:直接error{}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		//0参数是this
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("service.registermethod: %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	//自定义参数类型是否可以导出或者时内置的类型
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	//对numCalls 原子操作
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	//函数调用,
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
