package di

import "reflect"

type ItemFactory func(Container) any

type ItemDescriptor struct {
	name     *string
	itemType reflect.Type
	lifetime Lifetime
	instance *reflect.Value
	factory  ItemFactory
}

func (des *ItemDescriptor) Name() *string {
	return des.name
}
func (des *ItemDescriptor) ItemType() reflect.Type {
	return des.itemType
}
func (des *ItemDescriptor) Lifetime() Lifetime {
	return des.lifetime
}
func (des *ItemDescriptor) Instance() any {
	return des.instance
}
func (des *ItemDescriptor) Factory() ItemFactory {
	return des.factory
}
