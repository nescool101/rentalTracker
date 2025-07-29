package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	supa "github.com/supabase-community/supabase-go"

	"github.com/nescool101/rentManager/model"
)

// PropertyRepository provides methods to interact with the Property table in Supabase
type PropertyRepository struct {
	client *supa.Client
}

// NewPropertyRepository creates a new PropertyRepository
func NewPropertyRepository(client *supa.Client) *PropertyRepository {
	return &PropertyRepository{
		client: client,
	}
}

// GetManagerIDsForProperty retrieves all manager person IDs for a given property ID.
func (r *PropertyRepository) GetManagerIDsForProperty(ctx context.Context, propertyID uuid.UUID) ([]uuid.UUID, error) {
	var results []struct {
		ManagerPersonID uuid.UUID `json:"manager_person_id"`
	}

	data, _, err := r.client.From("property_managers").
		Select("manager_person_id", "exact", false).
		Eq("property_id", propertyID.String()).
		Execute()

	if err != nil {
		log.Printf("Error fetching manager IDs for property %s: %v", propertyID, err)
		return nil, fmt.Errorf("failed to fetch manager IDs: %w", err)
	}

	if err := json.Unmarshal(data, &results); err != nil {
		// It's possible that 'data' is empty if no managers are linked, which is not an unmarshal error.
		// Check if data is effectively empty or just '[]'
		if string(data) != "" && string(data) != "[]" {
			log.Printf("Error unmarshaling manager IDs for property %s: %v. Data: %s", propertyID, err, string(data))
			return nil, fmt.Errorf("failed to parse manager IDs: %w", err)
		}
		// If data is empty or [], it means no managers, so return empty slice
		return []uuid.UUID{}, nil
	}

	managerIDs := make([]uuid.UUID, len(results))
	for i, res := range results {
		managerIDs[i] = res.ManagerPersonID
	}
	return managerIDs, nil
}

// GetAll retrieves all properties from the database
func (r *PropertyRepository) GetAll(ctx context.Context) ([]model.Property, error) {
	var properties []model.Property

	data, count, err := r.client.From("property").Select("*", "exact", false).Execute()
	if err != nil {
		log.Printf("Error fetching properties: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d properties", count)

	err = json.Unmarshal([]byte(data), &properties)
	if err != nil {
		log.Printf("Error parsing property data: %v", err)
		return nil, err
	}

	// Populate ManagerIDs for each property
	for i := range properties {
		p := &properties[i]
		managerIDs, managerErr := r.GetManagerIDsForProperty(ctx, p.ID)
		if managerErr != nil {
			log.Printf("Error fetching manager IDs for property %s during GetAll: %v", p.ID, managerErr)
			// Continue, property will have nil ManagerIDs
		}
		p.ManagerIDs = managerIDs
	}

	return properties, nil
}

// GetByID retrieves a property by ID and populates its ManagerIDs.
func (r *PropertyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Property, error) {
	data, dbCount, err := r.client.From("property").Select("*", "exact", false).
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error fetching property by ID %s: %v", id, err)
		return nil, err
	}

	if dbCount == 0 {
		return nil, nil // Not found
	}

	var properties []model.Property // Fetch as a slice first
	if err := json.Unmarshal(data, &properties); err != nil {
		log.Printf("Error parsing property data for ID %s: %v", id, err)
		return nil, err
	}

	if len(properties) == 0 {
		return nil, nil // Should not happen if dbCount > 0, but defensive
	}

	property := &properties[0]

	// Populate ManagerIDs
	managerIDs, managerErr := r.GetManagerIDsForProperty(ctx, property.ID)
	if managerErr != nil {
		log.Printf("Error fetching manager IDs for property %s during GetByID: %v", property.ID, managerErr)
		// Continue, property will have nil ManagerIDs if this fails
	}
	property.ManagerIDs = managerIDs

	return property, nil
}

// GetByResident retrieves properties by resident ID
func (r *PropertyRepository) GetByResident(ctx context.Context, residentID uuid.UUID) ([]model.Property, error) {
	var properties []model.Property
	data, count, err := r.client.From("property").Select("*", "exact", false).
		Eq("resident_id", residentID.String()).Execute()
	if err != nil {
		log.Printf("Error fetching properties by resident ID %s: %v", residentID, err)
		return nil, err
	}

	if count == 0 {
		return []model.Property{}, nil // No properties found
	}

	err = json.Unmarshal([]byte(data), &properties)
	if err != nil {
		log.Printf("Error parsing property data for resident ID %s: %v", residentID, err)
		return nil, err
	}

	// Populate ManagerIDs for each property
	for i := range properties {
		p := &properties[i]
		managerIDs, managerErr := r.GetManagerIDsForProperty(ctx, p.ID)
		if managerErr != nil {
			log.Printf("Error fetching manager IDs for property %s during GetByResident: %v", p.ID, managerErr)
		}
		p.ManagerIDs = managerIDs
	}

	return properties, nil
}

// GetPropertiesForManager retrieves properties associated with the given manager_person_id.
func (r *PropertyRepository) GetPropertiesForManager(ctx context.Context, managerPersonID uuid.UUID) ([]model.Property, error) {
	var propertyManagerLinks []struct {
		PropertyID uuid.UUID `json:"property_id"`
	}

	pmData, _, err := r.client.From("property_managers").
		Select("property_id", "exact", false).
		Eq("manager_person_id", managerPersonID.String()).
		Execute()

	if err != nil {
		log.Printf("Error fetching property links for manager %s: %v", managerPersonID, err)
		return nil, fmt.Errorf("failed to fetch property links for manager: %w", err)
	}

	if err := json.Unmarshal(pmData, &propertyManagerLinks); err != nil {
		if string(pmData) != "" && string(pmData) != "[]" {
			log.Printf("Error unmarshaling property links for manager %s: %v. Data: %s", managerPersonID, err, string(pmData))
			return nil, fmt.Errorf("failed to parse property links for manager: %w", err)
		}
		return []model.Property{}, nil // No links found
	}

	if len(propertyManagerLinks) == 0 {
		return []model.Property{}, nil // No properties managed by this person
	}

	propertyIDs := make([]string, len(propertyManagerLinks))
	for i, link := range propertyManagerLinks {
		propertyIDs[i] = link.PropertyID.String()
	}

	var properties []model.Property
	propData, _, err := r.client.From("property").
		Select("*", "exact", false).
		In("id", propertyIDs).
		Execute()

	if err != nil {
		log.Printf("Error fetching properties by IDs for manager %s: %v", managerPersonID, err)
		return nil, fmt.Errorf("failed to fetch properties for manager: %w", err)
	}

	if err := json.Unmarshal(propData, &properties); err != nil {
		log.Printf("Error unmarshaling properties for manager %s: %v", managerPersonID, err)
		return nil, fmt.Errorf("failed to parse properties for manager: %w", err)
	}

	for i := range properties {
		p := &properties[i]
		managerIDs, managerErr := r.GetManagerIDsForProperty(ctx, p.ID)
		if managerErr != nil {
			log.Printf("Error fetching manager IDs for property %s during GetPropertiesForManager: %v", p.ID, managerErr)
		}
		p.ManagerIDs = managerIDs
	}

	return properties, nil
}

// AddManagerToProperty creates a link between a property and a manager.
func (r *PropertyRepository) AddManagerToProperty(ctx context.Context, propertyID uuid.UUID, managerPersonID uuid.UUID) error {
	_, _, err := r.client.From("property_managers").
		Insert(map[string]string{
			"property_id":       propertyID.String(),
			"manager_person_id": managerPersonID.String(),
		}, false, "exact", "", "ignore"). // "ignore" on conflict if the link already exists
		Execute()
	if err != nil {
		log.Printf("Error adding manager %s to property %s: %v", managerPersonID, propertyID, err)
		return fmt.Errorf("failed to add manager to property: %w", err)
	}
	return nil
}

// RemoveManagerFromProperty removes a link between a property and a manager.
func (r *PropertyRepository) RemoveManagerFromProperty(ctx context.Context, propertyID uuid.UUID, managerPersonID uuid.UUID) error {
	_, _, err := r.client.From("property_managers").
		Delete("exact", ""). // Use "exact" to ensure we get a count if needed, or "minimal"
		Eq("property_id", propertyID.String()).
		Eq("manager_person_id", managerPersonID.String()).
		Execute()
	if err != nil {
		log.Printf("Error removing manager %s from property %s: %v", managerPersonID, propertyID, err)
		return fmt.Errorf("failed to remove manager from property: %w", err)
	}
	return nil
}

// Create adds a new property to the database and links its managers.
func (r *PropertyRepository) Create(ctx context.Context, property model.Property) (*model.Property, error) {
	// Create a map for the property data, excluding ManagerIDs as it's not a direct column
	propertyData := map[string]interface{}{
		"id":          property.ID,
		"address":     property.Address,
		"apt_number":  property.AptNumber,
		"city":        property.City,
		"state":       property.State,
		"zip_code":    property.ZipCode,
		"type":        property.Type,
		"resident_id": property.ResidentID,
		// ManagerID is no longer here
	}

	data, _, err := r.client.From("property").Insert(propertyData, false, "exact", "", "").Execute()
	if err != nil {
		log.Printf("Error creating property: %v", err)
		return nil, err
	}

	var createdProperties []model.Property
	err = json.Unmarshal(data, &createdProperties)
	if err != nil {
		log.Printf("Error parsing created property data: %v", err)
		return nil, err
	}

	if len(createdProperties) == 0 {
		return nil, fmt.Errorf("failed to parse created property, empty result set")
	}

	createdProperty := &createdProperties[0]

	// Link managers
	if len(property.ManagerIDs) > 0 {
		for _, managerID := range property.ManagerIDs {
			if err := r.AddManagerToProperty(ctx, createdProperty.ID, managerID); err != nil {
				// Log error and continue, or decide if this should be a fatal error for the create operation
				log.Printf("Error linking manager %s to new property %s: %v", managerID, createdProperty.ID, err)
			}
		}
	}
	// Repopulate ManagerIDs from the database to ensure consistency
	currentManagerIDs, managerErr := r.GetManagerIDsForProperty(ctx, createdProperty.ID)
	if managerErr != nil {
		log.Printf("Error fetching manager IDs for newly created property %s: %v", createdProperty.ID, managerErr)
	}
	createdProperty.ManagerIDs = currentManagerIDs

	return createdProperty, nil
}

// Update updates an existing property and its manager links.
func (r *PropertyRepository) Update(ctx context.Context, property model.Property) (*model.Property, error) {
	// Update scalar fields of the property
	propertyData := map[string]interface{}{
		"address":     property.Address,
		"apt_number":  property.AptNumber,
		"city":        property.City,
		"state":       property.State,
		"zip_code":    property.ZipCode,
		"type":        property.Type,
		"resident_id": property.ResidentID,
	}

	_, _, err := r.client.From("property").Update(propertyData, "exact", "").
		Eq("id", property.ID.String()).Execute()
	if err != nil {
		log.Printf("Error updating property %s: %v", property.ID, err)
		return nil, fmt.Errorf("failed to update property: %w", err)
	}

	// Manage manager links
	currentManagerIDs, err := r.GetManagerIDsForProperty(ctx, property.ID)
	if err != nil {
		log.Printf("Error fetching current managers for property %s during update: %v", property.ID, err)
		return nil, fmt.Errorf("failed to fetch current managers: %w", err)
	}

	requestedManagerIDsMap := make(map[uuid.UUID]bool)
	for _, id := range property.ManagerIDs {
		requestedManagerIDsMap[id] = true
	}

	currentManagerIDsMap := make(map[uuid.UUID]bool)
	for _, id := range currentManagerIDs {
		currentManagerIDsMap[id] = true
	}

	// Add new managers
	for managerID := range requestedManagerIDsMap {
		if !currentManagerIDsMap[managerID] {
			if err := r.AddManagerToProperty(ctx, property.ID, managerID); err != nil {
				log.Printf("Error adding manager %s to property %s during update: %v", managerID, property.ID, err)
				// Potentially collect errors and return them, or handle as critical
			}
		}
	}

	// Remove old managers
	for managerID := range currentManagerIDsMap {
		if !requestedManagerIDsMap[managerID] {
			if err := r.RemoveManagerFromProperty(ctx, property.ID, managerID); err != nil {
				log.Printf("Error removing manager %s from property %s during update: %v", managerID, property.ID, err)
				// Potentially collect errors
			}
		}
	}

	// Fetch the updated property to return it with all fields populated correctly
	// (including any changes to manager IDs from the operations above)
	return r.GetByID(ctx, property.ID)
}

// Delete removes a property from the database
func (r *PropertyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// The property_managers table should have ON DELETE CASCADE for property_id
	// so manager links will be removed automatically when the property is deleted.
	_, _, err := r.client.From("property").Delete("minimal", "").
		Eq("id", id.String()).Execute()
	if err != nil {
		log.Printf("Error deleting property %s: %v", id, err)
		return fmt.Errorf("failed to delete property: %w", err)
	}

	return nil
}

// GetByUserID retrieves properties that a user is renting
func (r *PropertyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Property, error) {
	rentalsData, count, err := r.client.From("rental").Select("property_id", "exact", false).
		Eq("renter_id", userID.String()).Execute()
	if err != nil {
		log.Printf("Error fetching rentals for user %s: %v", userID.String(), err)
		return nil, err
	}

	if count == 0 {
		return []model.Property{}, nil
	}

	var rentals []struct {
		PropertyID uuid.UUID `json:"property_id"`
	}
	err = json.Unmarshal([]byte(rentalsData), &rentals)
	if err != nil {
		log.Printf("Error parsing rental data for user %s: %v", userID.String(), err)
		return nil, err
	}

	var propertyIDs []string
	for _, rental := range rentals {
		propertyIDs = append(propertyIDs, rental.PropertyID.String())
	}

	if len(propertyIDs) == 0 {
		return []model.Property{}, nil
	}

	var allProperties []model.Property
	// Supabase might have limits on URL length for IN clause, process in batches if necessary
	// For simplicity, assuming the number of properties per user is not excessively large.
	propertiesData, _, err := r.client.From("property").Select("*", "exact", false).
		In("id", propertyIDs).Execute()

	if err != nil {
		log.Printf("Error fetching properties by IDs for user %s: %v", userID.String(), err)
		return nil, err
	}

	err = json.Unmarshal([]byte(propertiesData), &allProperties)
	if err != nil {
		log.Printf("Error parsing properties for user %s: %v", userID.String(), err)
		return nil, err
	}

	// Populate ManagerIDs for each property
	for i := range allProperties {
		p := &allProperties[i]
		managerIDs, managerErr := r.GetManagerIDsForProperty(ctx, p.ID)
		if managerErr != nil {
			log.Printf("Error fetching manager IDs for property %s during GetByUserID: %v", p.ID, managerErr)
		}
		p.ManagerIDs = managerIDs
	}

	return allProperties, nil
}
