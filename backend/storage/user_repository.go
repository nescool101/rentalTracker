package storage

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	supa "github.com/supabase-community/supabase-go"

	"github.com/nescool101/rentManager/model"
)

// UserRepository provides methods to interact with the users table in Supabase
type UserRepository struct {
	client *supa.Client
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(client *supa.Client) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

// GetAll retrieves all users from the database
func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	var users []model.User

	data, count, err := r.client.From("users").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d users", count)

	// Parse the JSON response data into our struct
	err = json.Unmarshal([]byte(data), &users)
	if err != nil {
		log.Printf("Error parsing user data: %v", err)
		return nil, err
	}

	return users, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	data, count, err := r.client.From("users").Select("*", "exact", false).
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error fetching user by ID: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	// Parse the first result
	var users []model.User
	err = json.Unmarshal([]byte(data), &users)
	if err != nil {
		log.Printf("Error parsing user data: %v", err)
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil // Not found
	}

	return &users[0], nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	data, count, err := r.client.From("users").Select("*", "exact", false).
		Eq("email", email).Execute()
	if err != nil {
		log.Printf("Error fetching user by email: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	// Parse the first result
	var users []model.User
	err = json.Unmarshal([]byte(data), &users)
	if err != nil {
		log.Printf("Error parsing user data: %v", err)
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil // Not found
	}

	return &users[0], nil
}

// GetByPersonID retrieves a user by PersonID
func (r *UserRepository) GetByPersonID(ctx context.Context, personID uuid.UUID) (*model.User, error) {
	data, count, err := r.client.From("users").Select("*", "exact", false).
		Eq("person_id", personID.String()).Execute()
	if err != nil {
		log.Printf("Error fetching user by PersonID %s: %v", personID, err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	// Parse the first result
	var users []model.User
	err = json.Unmarshal([]byte(data), &users)
	if err != nil {
		log.Printf("Error parsing user data for PersonID %s: %v", personID, err)
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil // Not found
	}

	return &users[0], nil
}

// Create adds a new user to the database
func (r *UserRepository) Create(ctx context.Context, user model.User) (*model.User, error) {
	data, count, err := r.client.From("users").Insert(user, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // No result returned
	}

	var users []model.User
	err = json.Unmarshal([]byte(data), &users)
	if err != nil {
		log.Printf("Error parsing user data: %v", err)
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil // No result returned
	}

	return &users[0], nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user model.User) (*model.User, error) {
	data, count, err := r.client.From("users").Update(user, "exact", "").
		Eq("id", user.ID.String()).Execute()
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // No result returned
	}

	var users []model.User
	err = json.Unmarshal([]byte(data), &users)
	if err != nil {
		log.Printf("Error parsing user data: %v", err)
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil // No result returned
	}

	return &users[0], nil
}

// Delete removes a user from the database
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, _, err := r.client.From("users").Delete("minimal", "").
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		return err
	}

	return nil
}
