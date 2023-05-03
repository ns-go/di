package di

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/ns-go/di/internal/utils"
)

type Container struct {
	items           []*ItemDescriptor
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

	var instance any
	if d.factory != nil {
		instance = d.factory(*c)
		if instance == nil {
			return nil, nil
		}
	} else {
		if d.itemType == nil {
			return nil, errors.New("Cannot create instance, Because unknow type of item.")
		}
		instance = reflect.New(d.itemType)
	}

	typeOfInstance := reflect.TypeOf(instance)
	if typeOfInstance.Kind() == reflect.Ptr {
		typeOfInstance = typeOfInstance.Elem()
	}

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

		// if f.fieldType.Kind() != reflect.Ptr {
		// 	return nil, errors.New("'di.inject' tag support only 'Pointer' type.")
		// }

		var des *ItemDescriptor
		if f.itemName != nil && *f.itemName != "" {
			descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool { return f.itemName == d.name })
			if len(descriptors) < 1 {
				return nil, errors.New(fmt.Sprintf("No any instance register with name '%s'.", *f.itemName))
			}

			des = descriptors[0]
			val, err := c.resolveItemValue(des)
			if err != nil {
				return nil, err
			}

			if f.fieldType != des.itemType {
				return nil, errors.New(fmt.Sprintf("Field '%s' type not match to item '%s'.", f.fieldName, *f.itemName))
			}

			reflect.ValueOf(instance).FieldByName(f.fieldName).Addr().Set(reflect.ValueOf(val))

		} else {
			descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool { return f.fieldType == d.itemType && (d.name == nil || *d.name == "") })
			if len(descriptors) < 1 {
				return nil, errors.New(fmt.Sprintf("Type '%s' dose not registered.", f.fieldType))
			}

			des = descriptors[0]
			val, err := c.resolveItemValue(des)
			if err != nil {
				return nil, err
			}
			if f.fieldType != des.itemType {
				return nil, errors.New(fmt.Sprintf("Field '%s' type not match to item '%s'.", f.fieldName, *f.itemName))
			}

			reflect.ValueOf(instance).FieldByName(f.fieldName).Addr().Set(reflect.ValueOf(val))
		}
	}

	return instance, nil
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
	descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool { return *d.name == name })
	if len(descriptors) == 0 {
		return nil, errors.New(fmt.Sprintf("No any instance register with name '%s'.", name))
	}
	val, err := c.resolveItemValue(descriptors[0])
	return val, err
}

func (c *Container) ResolveByType(t reflect.Type) (any, error) {
	descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool {
		return d.itemType != nil && d.itemType == t && (d.name == nil || *d.name == "")
	})
	if len(descriptors) == 0 {
		return nil, errors.New(fmt.Sprintf("Type '%s' not registered.", t))
	}

	val, err := c.resolveItemValue(descriptors[0])
	return val, err
}

func (c *Container) NewScope() (*Container, error) {
	if c.scoped {
		return nil, errors.New("Cannot create scope from none-master container.")
	}
	childContainer := Container{}
	childContainer.masterContainer = c
	childContainer.scoped = true
	items := make([]*ItemDescriptor, len(c.items))
	for i := 0; i < len(c.items); i++ {
		des := c.items[i]
		if des.lifetime == Singleton {
			items[i] = des
		} else {
			items[i] = &ItemDescriptor{
				name:     des.name,
				itemType: des.itemType,
				lifetime: des.lifetime,
				instance: nil,
				factory:  des.factory,
			}
		}
	}

	childContainer.items = items

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
	val, err := c.ResolveByType(reflect.TypeOf(new(TResult)))
	if err != nil {
		return nil, err
	}
	val2 := val.(TResult)
	return &val2, err
}

func RegisterScoped[T any](c *Container, safe bool) error {
	t := reflect.TypeOf(new(T))
	descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool {
		return d.itemType != nil && d.itemType == t && (d.name == nil || *d.name == "")
	})

	if len(descriptors) > 0 {
		err := errors.New(fmt.Sprintf("Type '%s' is already registered.", t))
		if safe {
			return err
		} else {
			panic(err)
		}
	} else {
		c.items = append(c.items, &ItemDescriptor{itemType: t, lifetime: Scoped})
	}

	return nil
}

func RegisterTransient[T any](c *Container, safe bool) error {
	t := reflect.TypeOf(new(T))
	descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool {
		return d.itemType != nil && d.itemType == t && (d.name == nil || *d.name == "")
	})

	if len(descriptors) > 0 {
		err := errors.New(fmt.Sprintf("Type '%s' is already registered.", t))
		if safe {
			return err
		} else {
			panic(err)
		}
	} else {
		c.items = append(c.items, &ItemDescriptor{itemType: t, lifetime: Transient})
	}

	return nil
}

func RegisterSingleton[T any](c *Container, safe bool) error {
	t := reflect.TypeOf(new(T))
	descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool {
		return d.itemType != nil && d.itemType == t && (d.name == nil || *d.name == "")
	})

	if len(descriptors) > 0 {
		err := errors.New(fmt.Sprintf("Type '%s' is already registered.", t))
		if safe {
			return err
		} else {
			panic(err)
		}
	} else {
		c.items = append(c.items, &ItemDescriptor{itemType: t, lifetime: Singleton})
	}

	return nil
}

func RegisterByName(c *Container, name string, value any, safe bool) error {
	if value == nil {
		err := errors.New("Value could not be null.")
		if safe {
			return err
		}
	}
	t := reflect.TypeOf(value)
	descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool {
		return d.name != nil && *d.name == name
	})

	if len(descriptors) > 0 {
		err := errors.New(fmt.Sprintf("Item name '%s' is already registered.", name))
		if safe {
			return err
		} else {
			panic(err)
		}
	} else {
		c.items = append(c.items, &ItemDescriptor{itemType: t, lifetime: Singleton, instance: value})
	}

	return nil
}

func RegisterFactory[T any](c *Container, lifetime Lifetime, factory func(Container) T, safe bool) error {
	if factory == nil {
		err := errors.New("Factory could not be null.")
		if safe {
			return err
		}
	}

	t := reflect.TypeOf(new(T))
	descriptors := utils.FilterSlice(c.items, func(d *ItemDescriptor) bool {
		return d.itemType != nil && d.itemType == t && (d.name == nil || *d.name == "")
	})

	if len(descriptors) > 0 {
		err := errors.New(fmt.Sprintf("Type '%s' is already registered.", t))
		if safe {
			return err
		} else {
			panic(err)
		}
	} else {
		var fac ItemFactory = func(c Container) any { return factory(c) }
		c.items = append(c.items, &ItemDescriptor{itemType: t, lifetime: lifetime, factory: fac})
	}

	return nil
}

func NewContainer() *Container {
	return &Container{
		items:           make([]*ItemDescriptor, 0),
		scoped:          false,
		masterContainer: nil,
	}
}
