package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Multi-Site Models ---

type Site struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Slug      string             `bson:"slug" json:"slug"`
	Type      string             `bson:"type" json:"type"`
	Domains   []string           `bson:"domains" json:"domains"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type SiteMember struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID    primitive.ObjectID `bson:"site_id" json:"site_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Role      string             `bson:"role" json:"role"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// --- Singleton Documents (one document per site per collection) ---

type SiteSettings struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID          primitive.ObjectID `bson:"site_id" json:"site_id"`
	SiteTitle       string             `bson:"site_title" json:"site_title"`
	PageTitle       string             `bson:"page_title" json:"page_title"`
	MetaDescription string             `bson:"meta_description" json:"meta_description"`
	FooterTagline   string             `bson:"footer_tagline" json:"footer_tagline"`
	DefaultTheme    string             `bson:"default_theme" json:"default_theme"`
	ProfileImage    string             `bson:"profile_image" json:"profile_image"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type Hero struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID           primitive.ObjectID `bson:"site_id" json:"site_id"`
	Greeting         string             `bson:"greeting" json:"greeting"`
	FullName         string             `bson:"full_name" json:"full_name"`
	Subtitle         string             `bson:"subtitle" json:"subtitle"`
	CTAPrimaryText   string             `bson:"cta_primary_text" json:"cta_primary_text"`
	CTAPrimaryLink   string             `bson:"cta_primary_link" json:"cta_primary_link"`
	CTASecondaryText string             `bson:"cta_secondary_text" json:"cta_secondary_text"`
	CTASecondaryLink string             `bson:"cta_secondary_link" json:"cta_secondary_link"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

type Stat struct {
	Value string `bson:"value" json:"value"`
	Label string `bson:"label" json:"label"`
}

type About struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID          primitive.ObjectID `bson:"site_id" json:"site_id"`
	Title           string             `bson:"title" json:"title"`
	BioParagraphs   []string           `bson:"bio_paragraphs" json:"bio_paragraphs"`
	PersonalityTags []string           `bson:"personality_tags" json:"personality_tags"`
	Stats           []Stat             `bson:"stats" json:"stats"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// --- Collection Documents (multiple documents per collection) ---

type Skill struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID    primitive.ObjectID `bson:"site_id" json:"site_id"`
	Name      string             `bson:"name" json:"name"`
	Icon      string             `bson:"icon" json:"icon"`
	Category  string             `bson:"category" json:"category"`
	SortOrder int                `bson:"sort_order" json:"sort_order"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID      primitive.ObjectID `bson:"site_id" json:"site_id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Tags        []string           `bson:"tags" json:"tags"`
	Image       string             `bson:"image,omitempty" json:"image,omitempty"`
	LiveURL     string             `bson:"live_url,omitempty" json:"live_url,omitempty"`
	SourceURL   string             `bson:"source_url,omitempty" json:"source_url,omitempty"`
	SortOrder   int                `bson:"sort_order" json:"sort_order"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type Experience struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID      primitive.ObjectID `bson:"site_id" json:"site_id"`
	Role        string             `bson:"role" json:"role"`
	Company     string             `bson:"company" json:"company"`
	Period      string             `bson:"period" json:"period"`
	Description string             `bson:"description" json:"description"`
	Highlights  []string           `bson:"highlights" json:"highlights"`
	SortOrder   int                `bson:"sort_order" json:"sort_order"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type SocialLink struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID    primitive.ObjectID `bson:"site_id" json:"site_id"`
	Name      string             `bson:"name" json:"name"`
	URL       string             `bson:"url" json:"url"`
	Icon      string             `bson:"icon" json:"icon"`
	SortOrder int                `bson:"sort_order" json:"sort_order"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type ContactMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SiteID    primitive.ObjectID `bson:"site_id" json:"site_id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Subject   string             `bson:"subject" json:"subject"`
	Message   string             `bson:"message" json:"message"`
	IsRead    bool               `bson:"is_read" json:"is_read"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleEditor     = "editor"
	RoleViewer     = "viewer"
)

var ValidRoles = map[string]int{
	RoleViewer:     1,
	RoleEditor:     2,
	RoleAdmin:      3,
	RoleSuperAdmin: 4,
}

func RoleLevel(role string) int {
	return ValidRoles[role]
}

const (
	SiteRoleOwner  = "owner"
	SiteRoleEditor = "editor"
	SiteRoleViewer = "viewer"
)

var ValidSiteRoles = map[string]int{
	SiteRoleViewer: 1,
	SiteRoleEditor: 2,
	SiteRoleOwner:  3,
}

func SiteRoleLevel(role string) int {
	return ValidSiteRoles[role]
}

const (
	SiteTypePortfolio = "portfolio"
	SiteTypeShop      = "shop"
	SiteTypeFinance   = "finance"
)

var ValidSiteTypes = map[string]bool{
	SiteTypePortfolio: true,
	SiteTypeShop:      true,
	SiteTypeFinance:   true,
}

type AdminUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"-"`
	Role      string             `bson:"role" json:"role"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// --- API Request/Response Types ---

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string    `json:"token"`
	User  LoginUser `json:"user"`
}

type LoginUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type ReorderRequest struct {
	IDs []string `json:"ids"`
}

type CreateSiteRequest struct {
	Name    string   `json:"name"`
	Slug    string   `json:"slug"`
	Type    string   `json:"type"`
	Domains []string `json:"domains"`
}

type UpdateSiteRequest struct {
	Name    string   `json:"name"`
	Slug    string   `json:"slug"`
	Domains []string `json:"domains"`
}

type AddSiteMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type UpdateSiteMemberRequest struct {
	Role string `json:"role"`
}

type UserMembershipAssignment struct {
	SiteID string `json:"site_id"`
}

type UpdateUserMembershipsRequest struct {
	Memberships []UserMembershipAssignment `json:"memberships"`
}

type UserMembershipResponse struct {
	ID     string `json:"id"`
	SiteID string `json:"site_id"`
	UserID string `json:"user_id"`
}

type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type PortfolioData struct {
	SiteSettings SiteSettings  `json:"site_settings"`
	Hero         Hero          `json:"hero"`
	About        About         `json:"about"`
	Skills       []Skill       `json:"skills"`
	Projects     []Project     `json:"projects"`
	Experiences  []Experience  `json:"experiences"`
	SocialLinks  []SocialLink  `json:"social_links"`
	NavItems     []NavItem     `json:"nav_items"`
}

type NavItem struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}
