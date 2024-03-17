package main

import (
	"fmt"
	"reflect"
)

type People interface {
	ShowName() string
}

type User struct {
	Name     string `json:"name"`
	Age      int
	nickname string
}

func (u *User) Hello() string {
	return fmt.Sprintf("hello %s", u.Name)
}

func (u *User) Hello2() string {
	return fmt.Sprintf("hello2 %s", u.Name)
}

func (u User) HelloNickname() string {
	fmt.Println("enter nickname")
	return fmt.Sprintf("hello %s", u.nickname)
}

func (u User) HelloNickname2() string {
	return fmt.Sprintf("hello2 %s", u.nickname)
}

type User2 struct {
	Name string
}

func (u *User2) ShowName() string {
	return u.Name
}

func Add(a, b int) int {
	return a + b
}

func main() {
	//typeOf()
	//valueOf()
	//valueOf2()
	newObj()
}

func typeOf() {
	typeI := reflect.TypeOf(2)
	typeS := reflect.TypeOf("hello")
	fmt.Println(typeI) // int
	fmt.Println(typeS) // string
	fmt.Println()

	ptrUser := reflect.TypeOf(&User{})
	fmt.Println(ptrUser)               // *mian.User
	fmt.Println(ptrUser.Kind())        // prt
	fmt.Println(ptrUser.Elem())        // main.User
	fmt.Println(ptrUser.Elem().Kind()) // struct
	fmt.Println()

	typeUser := reflect.TypeOf(User{})
	fmt.Println(typeUser)        // main.User
	fmt.Println(typeUser.Kind()) // struct
	//fmt.Println(typeUser.Elem()) // panic
	fmt.Println()

	// 指针 转为 非指针类型
	fmt.Println(ptrUser.Elem() == typeUser) // true
	fmt.Println()

	// 获取 struct 成员变量的信息
	// 需要用 struct 的 Type，不能使用指针的 Type
	fieldNum := typeUser.NumField() // 成员变量个数
	for i := 0; i < fieldNum; i++ {
		field := typeUser.Field(i)
		fmt.Printf("%d %s | offset %d | anonymous %t | type %s | exported %t | json tag %s\n",
			i,
			field.Name,            // 变量名
			field.Offset,          // 相对于结构体首地址的内存偏移量，string 类型占据 16 个字节
			field.Anonymous,       // 是否为匿名成员
			field.Type,            //	类型
			field.IsExported(),    // 是否包外可见
			field.Tag.Get("json"), // 获取变量后面 `` 里面定义的 tag
		)
	}
	fmt.Println()

	// 通过 FieldByName 获取 Field
	if nameField, ok := typeUser.FieldByName("Name"); ok {
		fmt.Printf("Name is exported %t \n", nameField.IsExported())
	}
	fmt.Println()

	// 获取 struct 成员方法的信息
	methodNum := typeUser.NumMethod()
	for i := 0; i < methodNum; i++ {
		method := typeUser.Method(i)
		fmt.Printf("method name: %s, type: %s, exported: %t \n", method.Name, method.Type, method.IsExported())
	}
	fmt.Println()

	// 成员方法的个数，接受者为指针或值的方法【都】包含在内
	methodNum = ptrUser.NumMethod()
	for i := 0; i < methodNum; i++ {
		method := ptrUser.Method(i)
		fmt.Printf("method name: %s, type: %s, exported: %t \n", method.Name, method.Type, method.IsExported())
	}
	fmt.Println()

	// 获取函数信息
	typeFunc := reflect.TypeOf(Add)
	fmt.Printf("is function type: %t\n", typeFunc.Kind() == reflect.Func)
	argInNum := typeFunc.NumIn()   // 入参个数
	argOutNum := typeFunc.NumOut() // 输出个数
	for i := 0; i < argInNum; i++ {
		argType := typeFunc.In(i)
		fmt.Printf("第 %d 个输入的类型 %s\n", i, argType)
	}
	for i := 0; i < argOutNum; i++ {
		argType := typeFunc.Out(i)
		fmt.Printf("第 %d 个输出的类型 %s\n", i, argType)
	}
	fmt.Println()

	// 判断类型是否实现了某接口
	// 通过 reflect.TypeOf((*<interface>)(nil)).Elem() 获得接口类型。因为接口不能创建实例，所以 nil 强转为接口类型
	typeOfPeople := reflect.TypeOf((*People)(nil)).Elem()
	fmt.Printf("typeOfPeople kind is interface %t\n", typeOfPeople.Kind() == reflect.Interface)
	u2 := reflect.TypeOf(&User2{})
	// 如果值类型实现了接口，则指针类型也实现了该接口；反之，不成立
	fmt.Printf("%s implements People interface %t\n", typeUser, typeUser.Implements(typeOfPeople))
	fmt.Printf("%s implements People interface %t\n", u2, u2.Implements(typeOfPeople))
}

func valueOf() {
	iVal := reflect.ValueOf(2)
	sVal := reflect.ValueOf("hello")
	ptrUser := reflect.ValueOf(&User{
		Name:     "Tom",
		Age:      22,
		nickname: "汤姆",
	})
	fmt.Println(iVal)    // 2
	fmt.Println(sVal)    // hello
	fmt.Println(ptrUser) // &{Tom 22 汤姆 }
	fmt.Println()

	// value 转 Type
	iType := iVal.Type()       // int
	sType := sVal.Type()       // string
	userType := ptrUser.Type() // *main.User
	fmt.Println(iType)
	fmt.Println(sType)
	fmt.Println(userType)
	fmt.Println(iType.Kind() == reflect.Int, iVal.Kind() == reflect.Int, iType.Kind() == iVal.Kind())
	fmt.Println(sType.Kind() == reflect.String, sVal.Kind() == reflect.String, sType.Kind() == sVal.Kind())
	fmt.Println(userType.Kind() == reflect.Ptr, ptrUser.Kind() == reflect.Ptr, userType.Kind() == ptrUser.Kind())
	fmt.Println()

	// 指针 value 和非指针 value 互转
	userVal := ptrUser.Elem() // Elem() 指针类型转为对应的非指针类型
	ptrVal := userVal.Addr()  // Addr() 非指针类型转为指针类型
	fmt.Println(ptrUser.Kind(), userVal.Kind(), ptrVal.Kind())
	fmt.Println()

	// Value 对应的原始数据
	fmt.Printf("origin value iVal is %d %d\n", iVal.Interface().(int), iVal.Int())
	fmt.Printf("origin value sVal is %s %s\n", sVal.Interface().(string), sVal.String())
	u1 := ptrUser.Interface().(*User)
	fmt.Printf("name=%s age=%d nickname=%s\n", u1.Name, u1.Age, u1.nickname)
	u2 := userVal.Interface().(User)
	fmt.Printf("name=%s age=%d nickname=%s\n", u2.Name, u2.Age, u2.nickname)
	fmt.Println()

	// 空 Value 的判断
	var i any // 接口没有指向具体的值
	v := reflect.ValueOf(i)
	fmt.Printf("v 是否持有值：%t, v is Invalid：%t\n", v.IsValid(), v.Kind() == reflect.Invalid)
	fmt.Println()

	var user *User = nil
	pvu := reflect.ValueOf(user)
	if pvu.IsValid() {
		fmt.Printf("pvu 持有的值是对应类型的零值 %t\n", pvu.IsZero()) // 先判断 IsValid() 再调用 IsZero()，否则 panic
	}
	fmt.Println()

	var u User
	uv := reflect.ValueOf(u)
	if uv.IsValid() {
		fmt.Printf("uv 持有的值是对应类型的零值 %t\n", uv.IsZero()) // 先判断 IsValid() 再调用 IsZero()，否则 panic
	}
}

func valueOf2() {
	i, s := 10, "hello"
	user := User{"Tom", 22, "汤姆"}

	// golang 函数传参传的都是值，要修改原值需要传指针
	ptrI := reflect.ValueOf(&i)
	ptrS := reflect.ValueOf(&s)
	ptrUser := reflect.ValueOf(&user)
	// 通过指针的 Elem() 返回指针指向的对象
	ptrI.Elem().SetInt(8)
	ptrS.Elem().SetString("world")
	ptrUser.Elem().FieldByName("Age").SetInt(25)
	addrV := ptrUser.Elem().FieldByName("nickname")
	if addrV.CanSet() {
		addrV.SetString("Jack")
	} else {
		fmt.Println("nickname 是非可导出成员，不可Set")
	}
	fmt.Println(i)
	fmt.Println(s)
	fmt.Println(user)
	fmt.Println()

	// 修改切片
	users := make([]*User, 1, 5) // 要修改，用指针类型
	users[0] = &User{"Tom", 22, "汤姆"}

	sliceV := reflect.ValueOf(&users)
	if sliceV.Elem().Len() > 0 {
		sliceV.Elem().Index(0).Elem().FieldByName("Name").SetString("Helen")
		fmt.Println("1st user name changed")
		fmt.Println(users[0])
	}

	sliceV.Elem().SetCap(3)
	fmt.Println("cap", cap(users))
	// 增大len，向切片里追加元素
	sliceV.Elem().SetLen(2)
	sliceV.Elem().Index(1).Set(reflect.ValueOf(&User{"Jack", 30, "杰克"}))
	fmt.Printf("2st user name %s\n", users[1].Name)
	fmt.Println()

	// 修改 map
	u1 := User{"Tom", 22, "汤姆"}
	u2 := User{"Jack", 30, "杰克"}
	userM := make(map[string]*User, 5)
	userM[u1.Name] = &u1

	mapV := reflect.ValueOf(&userM)
	mapV.Elem().SetMapIndex(reflect.ValueOf(u2.Name), reflect.ValueOf(&u2))
	fmt.Println(len(userM))
	mapV.Elem().MapIndex(reflect.ValueOf(u1.Name)).Elem().FieldByName("Name").SetString("John")
	for k, u := range userM {
		fmt.Printf("key: %s, name: %s\n", k, u.Name)
	}
	fmt.Println()

	// 调用函数
	valueFn := reflect.ValueOf(Add) // 函数也是一种数据类型
	typeFn := reflect.TypeOf(Add)
	argInNum := typeFn.NumIn()
	args := make([]reflect.Value, argInNum) // 准备参数
	for i := 0; i < argInNum; i++ {
		if typeFn.In(i).Kind() == reflect.Int {
			args[i] = reflect.ValueOf(3) // 参数赋值
		}
	}
	rtVal := valueFn.Call(args) // 返回[]reflect.Value，因为 golang 可多参返回
	if typeFn.Out(0).Kind() == reflect.Int {
		rt := rtVal[0].Interface().(int)
		fmt.Printf("result=%d\n", rt)
	}
	fmt.Println()

	// 调用对象的成员方法
	userVal := reflect.ValueOf(&u1)
	helloMethod := userVal.MethodByName("Hello")
	rtVal = helloMethod.Call([]reflect.Value{}) // 无参的调用
	rtVal = helloMethod.Call(nil)               // 两种都可以
	rt := rtVal[0].Interface().(string)
	fmt.Printf("hello=%s\n", rt)

	// HelloNickname() 在定义时不是指针，userVal 是指针也可以调用
	helloNicknameMethod := userVal.MethodByName("HelloNickname")
	helloNicknameMethod.Call(nil)

	userVal2 := reflect.ValueOf(u1)
	helloNicknameMethod = userVal2.MethodByName("HelloNickname")
	helloNicknameMethod.Call(nil)
	//helloMethod = userVal2.MethodByName("Hello") // panic Hello 是指针类型的方法
	//rtVal = helloMethod.Call(nil)
}

func newObj() {
	u := reflect.TypeOf(User{})
	val := reflect.New(u) // 根据 reflect.Type 创建对象，得到对象的指针，再根据指针得到 reflect.Value
	val.Elem().FieldByName("Name").SetString("Test")
	user := val.Interface().(*User) // 把反射类型转成 go 原始类型
	fmt.Println(user)
	fmt.Println()

	// 切片
	var slice []User
	sliceType := reflect.TypeOf(slice)
	sliceValue := reflect.MakeSlice(sliceType, 1, 3)
	sliceValue.Index(0).Set(reflect.ValueOf(User{"Tom", 25, "汤姆"}))
	users := sliceValue.Interface().([]User)
	fmt.Printf("1st user name: %s\n", users[0].Name)
	fmt.Println()

	// map
	var userMap map[string]*User
	mType := reflect.TypeOf(userMap)
	mValue := reflect.MakeMap(mType)

	u1 := User{"Tom", 25, "汤姆"}
	key := reflect.ValueOf(u1.Name)
	mValue.SetMapIndex(key, reflect.ValueOf(&u1))
	mValue.MapIndex(key).Elem().FieldByName("Name").SetString("John")
	userMap = mValue.Interface().(map[string]*User)
	fmt.Printf("user name %s %s\n", userMap["Tom"].Name, u1.Name)
}
