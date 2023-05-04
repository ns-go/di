package di

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"unsafe"

	"github.com/ns-go/di/internal/utils"
)

type Container struct {
	namedItems      map[string]*ItemDescriptor
	typeItems       map[reflect.Type]*ItemDescriptor
	scoped          bool
	masterContainer *Container
}

type injectFieldInfo struct {
	fieldName string
	fieldType reflect.Type
	itemName  *string
}

func (c *Container) createInstance(d *ItemDescriptor) (any, error) {
	tagExp := regexp.MustCompile("di.inject:")

	var value reflect.Value
	if d.factory != nil {
		instance := d.factory(*c)
		if instance == nil {
			return nil, nil
		}
		value = reflect.ValueOf(instance)
	} else {
		if d.itemType == nil {
			return nil, errors.New("Cannot create instance, Because unknow type of item.")
		}
		value = reflect.New(d.itemType).Elem()
	}

	typeOfInstance := d.itemType

	numField := typeOfInstance.NumField()
	fields := make([]reflect.StructField, numField)

	for i := 0; i < numField; i++ {
		fields[i] = typeOfInstance.Field(i)
	}

	fields = utils.FilterSlice(fields, func(f reflect.StructField) bool { return tagExp.MatchString(string(f.Tag)) })

	injectFields := utils.MapSlice(fields, func(f reflect.StructField) injectFieldInfo {
		result := injectFieldInfo{}
		result.fieldName = f.Name
		result.fieldType = f.Type
		tagName := f.Tag.Get("di.inject")
		if tagName != "" {
			result.itemName = &tagName
		}
		return result
	})

	for _, f := range injectFields {

		var des *ItemDescriptor
		var finstance any
		var fvalue reflect.Value
		var err error
		if f.itemName != nil && *f.itemName != "" {
			des = c.namedItems[*f.itemName]
			finstance, err = c.ResolveByName(*f.itemName)
			fvalue = reflect.ValueOf(finstance)
			if err != nil {
				return nil, err
			}

		} else {

			fieldType := f.fieldType

			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			des = c.typeItems[fieldType]
			finstance, err = c.ResolveByType(fieldType)
			fvalue = reflect.ValueOf(finstance)
			if err != nil {
				return nil, err
			}

		}

		fieldType := f.fieldType
		if fieldType.Kind() == reflect.Ptr {
			ptrValue := reflect.New(fieldType.Elem())
			ptrValue.Elem().Set(fvalue)
			fieldType = fieldType.Elem()
			fvalue = ptrValue
		}

		if fieldType != des.itemType {
			return nil, errors.New(fmt.Sprintf("Field '%s' type not match to item '%s'.", f.fieldName, *f.itemName))
		}

		f1 := value.FieldByName(f.fieldName)
		x := reflect.NewAt(f1.Type(), unsafe.Pointer(f1.UnsafeAddr())).Elem()
		x.Set(fvalue)

	}

	return value.Interface(), nil
}

func (c *Container) resolveItemValue(d *ItemDescriptor) (any, error) {
	if d.lifetime == Scoped && !c.scoped {
		return nil, errors.New("Cannot resolve scoped item with none scoped container.")
	}

	if d.lifetime == Singleton || d.lifetime == Scoped { //Scoped items are cloned from  master container
		if d.name != nil { //registered instant by name
			return d.instance, nil
		} else {
			if d.instance == nil {
				if ins, err := c.createInstance(d); err != nil {
					return nil, err
				} else {
					d.instance = ins
				}

				return d.instance, nil
			} else {
				return d.instance, nil
			}
		}
	} else {
		if ins, err := c.createInstance(d); err != nil {
			return nil, err
		} else {
			return ins, nil
		}
	}
}

func (c *Container) ResolveByName(name string) (any, error) {
	des := c.namedItems[name]
	if des == nil {
		return nil, errors.New(fmt.Sprintf("No any instance register by name '%s'.", name))
	}
	val, err := c.resolveItemValue(des)
	return val, err
}

func (c *Container) ResolveByType(t reflect.Type) (any, error) {
	des := c.typeItems[t]
	if des == nil {
		return nil, errors.New(fmt.Sprintf("Type '%s' not registered.", t.Name()))
	}

	val, err := c.resolveItemValue(des)
	return val, err
}

func (c *Container) NewScope() (*Container, error) {
	if c.scoped {
		return nil, errors.New("Cannot create scope from none-master container.")
	}
	childContainer := Container{}
	childContainer.masterContainer = c
	childContainer.scoped = true
	nameditems := make(map[string]*ItemDescriptor)
	typeitems := make(map[reflect.Type]*ItemDescriptor)

	for k, el := range c.namedItems {
		nameditems[k] = el
	}

	for k, el := range c.typeItems {
		if el.lifetime == Singleton {
			typeitems[k] = el
		} else {
			typeitems[k] = &ItemDescriptor{
				name:     el.name,
				itemType: el.itemType,
				lifetime: el.lifetime,
				instance: nil,
				factory:  el.factory,
			}
		}
	}

	childContainer.namedItems = nameditems
	childContainer.typeItems = typeitems

	return &childContainer, nil
}

func (c *Container) MasterContainer() *Container {
	return c.masterContainer
}

func ResolveByName[TResult any](c *Container, name string) (*TResult, error) {
	val, err := c.ResolveByName(name)
	if err != nil {
		return nil, err
	}
	val2 := val.(TResult)
	return &val2, err
}

func Resolve[TResult any](c *Container) (*TResult, error) {
	val, err := c.ResolveByType(reflect.TypeOf(new(TResult)).Elem())
	if err != nil {
		return nil, err
	}
	val2 := val.(TResult)
	return &val2, err
}

func (c *Container) RegisterType(t reflect.Type, lifetime Lifetime, safe bool) error {
	if t.Kind() == reflect.Ptr {
		err := errors.New("Cannot register type of pointer.")
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	des := c.typeItems[t]
	if des != nil {
		err := errors.New(fmt.Sprintf("Type '%s' is already registered.", t.Name()))
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	c.typeItems[t] = &ItemDescriptor{itemType: t, lifetime: lifetime}
	return nil
}

func (c *Container) RegisterByName(name string, value any, safe bool) error {
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Ptr {
		err := errors.New("Cannot register type of pointer.")
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	des := c.namedItems[name]
	if des != nil {
		err := errors.New(fmt.Sprintf("Item name '%s' is already registered.", name))
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	c.namedItems[name] = &ItemDescriptor{itemType: t, lifetime: Singleton, name: &name, instance: value}
	return nil
}

func (c *Container) RegisterFactory(t reflect.Type, lifetime Lifetime, factory ItemFactory, safe bool) error {
	if factory == nil {
		err := errors.New("Factory could not be null.")
		if safe {
			return err
		}
	}

	if t.Kind() == reflect.Ptr {
		err := errors.New("Cannot register type of pointer.")
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	des := c.typeItems[t]
	if des != nil {
		err := errors.New(fmt.Sprintf("Type '%s' is already registered.", t.Name()))
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	c.typeItems[t] = &ItemDescriptor{itemType: t, lifetime: lifetime, factory: factory}
	return nil
}

func RegisterScoped[T any](c *Container, safe bool) error {
	t := reflect.TypeOf(new(T)).Elem()
	err := c.RegisterType(t, Scoped, safe)
	return err
}

func RegisterTransient[T any](c *Container, safe bool) error {
	t := reflect.TypeOf(new(T)).Elem()
	err := c.RegisterType(t, Transient, safe)
	return err
}

func RegisterSingleton[T any](c *Container, safe bool) error {
	t := reflect.TypeOf(new(T)).Elem()
	err := c.RegisterType(t, Singleton, safe)
	return err
}

func RegisterByName(c *Container, name string, value any, safe bool) error {
	err := c.RegisterByName(name, value, safe)
	return err
}

func RegisterFactory[T any](c *Container, lifetime Lifetime, factory func(Container) T, safe bool) error {
	t := reflect.TypeOf(new(T)).Elem()
	err := c.RegisterFactory(t, lifetime, func(c Container) any { return factory(c) }, safe)

	return err
}

func NewContainer() *Container {
	return &Container{
		namedItems:      make(map[string]*ItemDescriptor),
		typeItems:       make(map[reflect.Type]*ItemDescriptor),
		scoped:          false,
		masterContainer: nil,
	}
}
