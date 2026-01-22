package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Factory is a function that creates an instance using the container
type Factory func(*Container) any

// Container manages dependency injection
type Container struct {
	services  sync.Map // Stores singleton instances
	factories sync.Map // Stores factory functions
	resolving sync.Map // Tracks dependencies being resolved (circular detection)
}

// NewContainer creates a new DI container
func NewContainer() *Container {
	return &Container{}
}

// Register registers a singleton instance
func (c *Container) Register(interfacePtr any, implementation any) {
	t := reflect.TypeOf(interfacePtr).Elem()
	c.services.Store(t, implementation)
}

// RegisterFactory registers a factory function for lazy initialization
func (c *Container) RegisterFactory(interfacePtr any, factory Factory) {
	t := reflect.TypeOf(interfacePtr).Elem()
	c.factories.Store(t, factory)
}

// Resolve retrieves or creates a service from the container
func (c *Container) Resolve(interfacePtr any) any {
	t := reflect.TypeOf(interfacePtr).Elem()

	// Check if already instantiated
	if service, ok := c.services.Load(t); ok {
		return service
	}

	// Check for circular dependency
	if _, isResolving := c.resolving.Load(t); isResolving {
		fmt.Printf("DI: Circular dependency detected for %v\n", t)
		return nil
	}

	// Mark as resolving
	c.resolving.Store(t, true)
	defer c.resolving.Delete(t)

	// Check for factory
	if factory, ok := c.factories.Load(t); ok {
		factoryFunc := factory.(Factory)

		// Call factory with container
		service := factoryFunc(c)
		if service == nil {
			fmt.Printf("DI: Factory returned nil for %v\n", t)
			return nil
		}

		// Cache the result
		c.services.Store(t, service)
		return service
	}

	fmt.Printf("DI: Failed to resolve %v - not registered\n", t)
	return nil
}

// IsRegistered checks if a service is registered
func (c *Container) IsRegistered(interfacePtr any) bool {
	t := reflect.TypeOf(interfacePtr).Elem()

	_, hasService := c.services.Load(t)
	_, hasFactory := c.factories.Load(t)

	return hasService || hasFactory
}

// Clear removes all registered services (useful for testing)
func (c *Container) Clear() {
	c.services = sync.Map{}
	c.factories = sync.Map{}
	c.resolving = sync.Map{}
}
