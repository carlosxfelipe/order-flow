package data

import (
	"sort"
	"sync"

	"order-flow/domain"
)

// Store provides thread-safe in-memory data storage pre-seeded with order-flow demo data.
type Store struct {
	mu     sync.RWMutex
	users  map[string]domain.User  // username -> User
	orders map[string]domain.Order // orderId -> Order
}

// NewStore initializes a new Store and runs the seeder.
func NewStore() *Store {
	s := &Store{
		users:  make(map[string]domain.User),
		orders: make(map[string]domain.Order),
	}
	s.seed()
	return s
}

func (s *Store) seed() {
	// 1. Seed Users (matches DbSeeder.cs)
	users := []domain.User{
		{ID: 1, Username: "admin", Password: "admin123", Role: "Admin"},
		{ID: 2, Username: "carlos", Password: "carlos123", Role: "Customer"},
		{ID: 3, Username: "maria", Password: "maria123", Role: "Customer"},
		{ID: 4, Username: "luis", Password: "luis123", Role: "Customer"},
	}
	for _, u := range users {
		s.users[u.Username] = u
	}

	// 2. Seed Orders (matches DbSeeder.cs)
	orders := []domain.Order{
		{
			ID:           "b1a2c3d4-0001-4000-8000-000000000001",
			UserID:       2,
			CustomerName: "Carlos Felipe",
			Status:       domain.StatusCreated,
			Items: []domain.OrderItem{
				{ID: "item_101", ProductName: "Coca-Cola Can 350ml", Price: 5.50, Quantity: 2},
				{ID: "item_102", ProductName: "Classic Burger", Price: 25.00, Quantity: 1},
			},
		},
		{
			ID:           "b1a2c3d4-0002-4000-8000-000000000002",
			UserID:       3,
			CustomerName: "Maria Isabelle",
			Status:       domain.StatusPaid,
			Items: []domain.OrderItem{
				{ID: "item_201", ProductName: "Coca-Cola 2L", Price: 12.00, Quantity: 1},
				{ID: "item_202", ProductName: "Pepperoni Pizza", Price: 45.00, Quantity: 1},
			},
		},
		{
			ID:           "b1a2c3d4-0003-4000-8000-000000000003",
			UserID:       4,
			CustomerName: "Luís Felipe",
			Status:       domain.StatusShipped,
			Items: []domain.OrderItem{
				{ID: "item_301", ProductName: "Coca-Cola Zero Can", Price: 5.50, Quantity: 1},
			},
		},
	}
	for _, o := range orders {
		s.orders[o.ID] = o
	}
}

// GetUserByUsername finds a user by username.
func (s *Store) GetUserByUsername(username string) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[username]
	return u, ok
}

// CreateUser creates a new user if the username is not already taken.
func (s *Store) CreateUser(username, password, role string) (domain.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; exists {
		return domain.User{}, false
	}

	maxID := 0
	for _, u := range s.users {
		if u.ID > maxID {
			maxID = u.ID
		}
	}

	newUser := domain.User{
		ID:       maxID + 1,
		Username: username,
		Password: password,
		Role:     role,
	}
	s.users[username] = newUser
	return newUser, true
}

// DeleteUser removes a user from the store by username.
func (s *Store) DeleteUser(username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[username]; !exists {
		return false
	}
	delete(s.users, username)
	return true
}

// GetOrder returns an order by its ID.
func (s *Store) GetOrder(id string) (domain.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	return o, ok
}

// ListOrders returns all orders, optionally filtered by user ID if not admin.
func (s *Store) ListOrders(userID int, isAdmin bool) []domain.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.Order
	for _, o := range s.orders {
		if isAdmin || o.UserID == userID || userID == 0 {
			result = append(result, o)
		}
	}
	// Sort by ID for deterministic outputs
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// SaveOrder creates or updates an order in the store.
func (s *Store) SaveOrder(o domain.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.ID] = o
}
