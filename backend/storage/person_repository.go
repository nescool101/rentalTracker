package storage

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/google/uuid"
	supa "github.com/supabase-community/supabase-go"

	"github.com/nescool101/rentManager/model"
)

// PersonRepository provides methods to interact with the Person table in Supabase
type PersonRepository struct {
	client *supa.Client
}

// NewPersonRepository creates a new PersonRepository
func NewPersonRepository(client *supa.Client) *PersonRepository {
	return &PersonRepository{
		client: client,
	}
}

// GetAll retrieves all persons from the database
func (r *PersonRepository) GetAll(ctx context.Context) ([]model.Person, error) {
	var persons []model.Person

	data, count, err := r.client.From("person").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching persons: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d persons", count)

	// Parse the JSON response data into our struct
	err = json.Unmarshal([]byte(data), &persons)
	if err != nil {
		log.Printf("Error parsing person data: %v", err)
		return nil, err
	}

	return persons, nil
}

// GetByID retrieves a person by ID
func (r *PersonRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Person, error) {
	data, count, err := r.client.From("person").Select("*", "exact", false).
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error fetching person by ID: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // Not found
	}

	// Parse the first result
	var persons []model.Person
	err = json.Unmarshal([]byte(data), &persons)
	if err != nil {
		log.Printf("Error parsing person data: %v", err)
		return nil, err
	}

	if len(persons) == 0 {
		return nil, nil // Not found
	}

	return &persons[0], nil
}

// Note: GetByEmail method was removed because email is no longer in the Person struct

// GetByRole retrieves persons by role
func (r *PersonRepository) GetByRole(ctx context.Context, roleName string) ([]model.Person, error) {
	// For a complex query like this with joins, we'll use an RPC function
	type Request struct {
		RoleName string `json:"role_name"`
	}

	// Call the RPC function
	data := r.client.Rpc("get_persons_by_role", "exact", Request{RoleName: roleName})
	// Check for errors
	if err := checkRPCError(data); err != nil {
		log.Printf("Error fetching persons by role: %v", err)
		return nil, err
	}

	var persons []model.Person
	err := json.Unmarshal([]byte(data), &persons)
	if err != nil {
		log.Printf("Error parsing person data: %v", err)
		return nil, err
	}

	return persons, nil
}

// Create adds a new person to the database
func (r *PersonRepository) Create(ctx context.Context, person model.Person) (*model.Person, error) {
	// The method signature for Insert is:
	// Insert(values interface{}, returning string, count string, onConflict string, ignoreDuplicates string)
	data, count, err := r.client.From("person").Insert(person, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating person: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // No result returned
	}

	var persons []model.Person
	err = json.Unmarshal([]byte(data), &persons)
	if err != nil {
		log.Printf("Error parsing person data: %v", err)
		return nil, err
	}

	if len(persons) == 0 {
		return nil, nil // No result returned
	}

	return &persons[0], nil
}

// Update updates an existing person
func (r *PersonRepository) Update(ctx context.Context, person model.Person) (*model.Person, error) {
	// The method signature for Update is:
	// Update(values interface{}, returning string, count string)
	data, count, err := r.client.From("person").Update(person, "exact", "").
		Eq("id", person.ID.String()).Execute()
	if err != nil {
		log.Printf("Error updating person: %v", err)
		return nil, err
	}

	if count == 0 {
		return nil, nil // No result returned
	}

	var persons []model.Person
	err = json.Unmarshal([]byte(data), &persons)
	if err != nil {
		log.Printf("Error parsing person data: %v", err)
		return nil, err
	}

	if len(persons) == 0 {
		return nil, nil // No result returned
	}

	return &persons[0], nil
}

// Delete removes a person from the database
func (r *PersonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, _, err := r.client.From("person").Delete("minimal", "").
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error deleting person: %v", err)
		return err
	}

	return nil
}

// GetByIDs retrieves multiple persons by their IDs
func (r *PersonRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Person, error) {
	if len(ids) == 0 {
		return []model.Person{}, nil
	}

	// Convert uuid.UUID slice to string slice for the query
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	var persons []model.Person
	// Use the .In(column, values) filter
	data, count, err := r.client.From("person").Select("*", "exact", false).
		In("id", stringIDs).Execute()

	if err != nil {
		log.Printf("Error fetching persons by IDs: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d persons by IDs", count)

	err = json.Unmarshal([]byte(data), &persons)
	if err != nil {
		log.Printf("Error parsing person data from GetByIDs: %v", err)
		return nil, err
	}

	return persons, nil
}

// checkRPCError is a utility function to check for errors in RPC responses
func checkRPCError(response string) error {
	// Simplistic error checking - in a real implementation, you would
	// parse the response and check for specific error structures
	if response == "" {
		return errors.New("empty response from RPC call")
	}

	// Check if the response contains an error message
	if json.Valid([]byte(response)) {
		var responseObj map[string]interface{}
		if err := json.Unmarshal([]byte(response), &responseObj); err == nil {
			if errorMsg, ok := responseObj["error"]; ok {
				return errors.New(errorMsg.(string))
			}
		}
	}

	return nil
}
