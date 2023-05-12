package di

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"unsafe"

	"github.com/ns-go/di/internal/utils"
)

// Container is the main dependency injection container.
type Container struct {
	namedItems      map[string]*ItemDescriptor
	typeItems       map[reflect.Type]*ItemDescriptor
	scoped          bool
	masterContainer *Container
}

// injectFieldInfo holds information about a struct field that requires injection.
type injectFieldInfo struct {
	fieldName string
	fieldType reflect.Type
	itemName  *string
}

// createInstance creates an instance of an item registered in the container.
func (c *Container) createInstance(d *ItemDescriptor) (*reflect.Value, error) {
	tagExp := regexp.MustCompile("di.inject:")

	var value reflect.Value

	// If the item has a factory function, call it to create the instance.
	if d.factory != nil {
		instance := d.factory(*c)
		if instance == nil {
			return nil, nil
		}
		value = reflect.ValueOf(instance)

		// If the factory function returns a non-nil value that is not a pointer,
		// create a new pointer to that value.
		if value.Kind() != reflect.Pointer {
			ptrVal := reflect.New(d.itemType)
			ptrVal.Elem().Set(value)
			value = ptrVal
		}
	} else {
		// If no factory function is defined, create a new instance using reflection.
		if d.itemType == nil {
			return nil, errors.New("cannot create instance, Because unknow type of item")
		}
		value = reflect.New(d.itemType)
	}

	typeOfInstance := d.itemType

	numField := typeOfInstance.NumField()
	fields := make([]reflect.StructField, numField)

	for i := 0; i < numField; i++ {
		fields[i] = typeOfInstance.Field(i)
	}

	// Filter the struct fields to include only those with the "di.inject" tag.
	fields = utils.FilterSlice(fields, func(f reflect.StructField) bool { return tagExp.MatchString(string(f.Tag)) })

	// Map the filtered fields to injectFieldInfo structs.
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

		// Check if the field type is a pointer.
		if f.fieldType.Kind() != reflect.Pointer {
			return nil, errors.New("type of injection field allow only pointer")
		}

		var des *ItemDescriptor
		var finstance *reflect.Value
		var err error

		// Resolve the dependency either by name or by type.
		if f.itemName != nil && *f.itemName != "" {
			des = c.namedItems[*f.itemName]
			finstance, err = c.resolveByName(*f.itemName)
			if err != nil {
				return nil, err
			}

		} else {

			fieldType := f.fieldType

			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			des = c.typeItems[fieldType]
			finstance, err = c.resolveByType(fieldType)
			if err != nil {
				return nil, err
			}

		}

		fieldType := f.fieldType

		if fieldType.Elem() != des.itemType {
			return nil, fmt.Errorf("field '%s' type not match to item '%s'", f.fieldName, *f.itemName)
		}

		var f1 reflect.Value
		if value.Kind() == reflect.Pointer {
			f1 = value.Elem().FieldByName(f.fieldName)
		} else {
			f1 = value.FieldByName(f.fieldName)
		}

		// Set the field value to the resolved instance.
		x := reflect.NewAt(f1.Type(), unsafe.Pointer(f1.UnsafeAddr())).Elem()
		x.Set((*finstance))

	}

	return &value, nil
}

func (c *Container) resolveItemValue(d *ItemDescriptor) (*reflect.Value, error) {
	if d.lifetime == Scoped && !c.scoped {
		return nil, errors.New("cannot resolve scoped item with none scoped container")
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

// resolveByName resolves an item from the container by name.
func (c *Container) resolveByName(name string) (*reflect.Value, error) {
	des := c.namedItems[name]
	if des == nil {
		return nil, fmt.Errorf("no any instance register by name '%s'", name)
	}
	val, err := c.resolveItemValue(des)
	return val, err
}

func (c *Container) ResolveByName(name string) (any, error) {
	des := c.namedItems[name]
	if des == nil {
		return nil, fmt.Errorf("no any instance register by name '%s'", name)
	}
	val, err := c.resolveItemValue(des)

	if err != nil {
		return nil, err
	}

	val1 := (*val).Interface()

	return val1, err
}

// resolveByType resolves an item from the container by type.
func (c *Container) resolveByType(t reflect.Type) (*reflect.Value, error) {
	des := c.typeItems[t]
	if des == nil {
		return nil, fmt.Errorf("type '%s' not registered", t.Name())
	}

	val, err := c.resolveItemValue(des)
	return val, err
}

func (c *Container) ResolveByType(t reflect.Type) (any, error) {
	des := c.typeItems[t]
	if des == nil {
		return nil, fmt.Errorf("type '%s' not registered", t.Name())
	}

	val, err := c.resolveItemValue(des)
	if err != nil {
		return nil, err
	}
	val1 := (*val).Interface()

	return val1, err
}

func (c *Container) NewScope() (*Container, error) {
	if c.scoped {
		return nil, errors.New("cannot create scope from none-master container")
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
	val2 := val.(*TResult)
	return val2, err
}

func Resolve[TResult any](c *Container) (*TResult, error) {
	val, err := c.resolveByType(reflect.TypeOf(new(TResult)).Elem())
	if err != nil {
		return nil, err
	}

	result := val.Interface().(*TResult)
	return result, err
}

func (c *Container) RegisterType(t reflect.Type, lifetime Lifetime, safe bool) error {
	if t.Kind() == reflect.Ptr {
		err := errors.New("cannot register type of pointer")
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	des := c.typeItems[t]
	if des != nil {
		err := fmt.Errorf("type '%s' is already registered", t.Name())
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	c.typeItems[t] = &ItemDescriptor{itemType: t, lifetime: lifetime}
	return nil
}

func (c *Container) RegisterValue(t reflect.Type, value any, safe bool) error {
	if t.Kind() == reflect.Ptr {
		err := errors.New("cannot register type of pointer")
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	var _t = t
	if t.Kind() == reflect.Pointer {
		_t = t.Elem()
	}
	des := c.typeItems[_t]
	if des != nil {
		err := fmt.Errorf("type '%s' is already registered", t.Name())
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	ptr := reflect.New(t)
	val := reflect.ValueOf(value)
	ptr.Elem().Set(val)
	c.typeItems[t] = &ItemDescriptor{itemType: t, lifetime: Singleton, instance: &ptr}

	if t.Kind() == reflect.Pointer {
		ptr := reflect.ValueOf(value)
		c.typeItems[_t] = &ItemDescriptor{itemType: t, lifetime: Singleton, instance: &ptr}
	} else {
		ptr := reflect.New(t)
		val := reflect.ValueOf(value)
		ptr.Elem().Set(val)
		c.typeItems[_t] = &ItemDescriptor{itemType: t, lifetime: Singleton, instance: &ptr}
	}
	return nil
}

func (c *Container) RegisterByName(name string, value any, safe bool) error {
	t := reflect.TypeOf(value)

	des := c.namedItems[name]
	if des != nil {
		err := fmt.Errorf("item name '%s' is already registered", name)
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	if t.Kind() == reflect.Pointer {
		ptr := reflect.ValueOf(value)
		c.namedItems[name] = &ItemDescriptor{itemType: t.Elem(), lifetime: Singleton, name: &name, instance: &ptr}
	} else {
		ptr := reflect.New(t)
		val := reflect.ValueOf(value)
		ptr.Elem().Set(val)
		c.namedItems[name] = &ItemDescriptor{itemType: t, lifetime: Singleton, name: &name, instance: &ptr}
	}

	return nil
}

func (c *Container) RegisterFactory(t reflect.Type, lifetime Lifetime, factory ItemFactory, safe bool) error {
	if factory == nil {
		err := errors.New("factory could not be null")
		if safe {
			return err
		}
	}

	if t.Kind() == reflect.Ptr {
		err := errors.New("cannot register type of pointer")
		if safe {
			return err
		} else {
			panic(err)
		}
	}

	des := c.typeItems[t]
	if des != nil {
		err := fmt.Errorf("type '%s' is already registered", t.Name())
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

func RegisterValue[T any](c *Container, value T, safe bool) error {
	t := reflect.TypeOf(new(T)).Elem()
	err := c.RegisterValue(t, value, safe)
	return err
}

func RegisterByName(c *Container, name string, value any, safe bool) error {
	err := c.RegisterByName(name, value, safe)
	return err
}

func RegisterFactory[T any](c *Container, lifetime Lifetime, factory func(Container) *T, safe bool) error {
	t := reflect.TypeOf(new(T)).Elem()
	err := c.RegisterFactory(t, lifetime, func(c Container) any { return factory(c) }, safe)

	return err
}

// NewContainer creates a new dependency injection container.
func NewContainer() *Container {
	return &Container{
		namedItems:      make(map[string]*ItemDescriptor),
		typeItems:       make(map[reflect.Type]*ItemDescriptor),
		scoped:          false,
		masterContainer: nil,
	}
}
