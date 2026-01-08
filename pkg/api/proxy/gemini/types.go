package gemini

// GenerateContentRequest represents Gemini API generate content request
type GenerateContentRequest struct {
	Contents       []Content       `json:"contents"`
	Tools          []Tool          `json:"tools,omitempty"`
	ToolConfig     *ToolConfig     `json:"toolConfig,omitempty"`
	SafetySettings []SafetySetting  `json:"safetySettings,omitempty"`
	SystemInstruction *Content      `json:"systemInstruction,omitempty"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

// Content represents content in Gemini format
type Content struct {
	Role  string `json:"role"`  // "user", "model", "function"
	Parts []Part `json:"parts"`
}

// Part represents a part of content
type Part struct {
	Text         string      `json:"text,omitempty"`
	InlineData   *InlineData `json:"inlineData,omitempty"`
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
}

// InlineData represents inline data (e.g., images)
type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// FunctionResponse represents a function response
type FunctionResponse struct {
	Name string `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// Tool represents a tool (e.g., function calling)
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations"`
}

// FunctionDeclaration represents a function declaration
type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolConfig represents tool configuration
type ToolConfig struct {
	FunctionCallingConfig *FunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

// FunctionCallingConfig represents function calling configuration
type FunctionCallingConfig struct {
	Mode                string   `json:"mode"`
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

// SafetySetting represents a safety setting
type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GenerationConfig represents generation configuration
type GenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	TopP           float64 `json:"topP,omitempty"`
	TopK           int     `json:"topK,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	StopSequences  []string `json:"stopSequences,omitempty"`
	ResponseMIMEType string `json:"responseMimeType,omitempty"`
}

// GenerateContentResponse represents Gemini API response
type GenerateContentResponse struct {
	Candidates      []Candidate `json:"candidates"`
	UsageMetadata   *UsageMetadata `json:"usageMetadata,omitempty"`
	ModelVersion    string `json:"modelVersion,omitempty"`
	PromptFeedback  *PromptFeedback `json:"promptFeedback,omitempty"`
}

// Candidate represents a generation candidate
type Candidate struct {
	Content       *Content        `json:"content,omitempty"`
	FinishReason  string          `json:"finishReason"` // "FINISH_REASON_UNSPECIFIED", "STOP", "MAX_TOKENS", "SAFETY", "RECITATION", "OTHER"
	Index         int             `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
	TokenLogProbs []TokenLogProb  `json:"tokenLogProbs,omitempty"`
	FinishMessage string          `json:"finishMessage,omitempty"`
}

// SafetyRating represents a safety rating
type SafetyRating struct {
	Category    string  `json:"category"`
	Probability string  `json:"probability"` // "HARM_PROBABILITY_UNSPECIFIED", "NEGLIGIBLE", "LOW", "MEDIUM", "HIGH"
	IsBlocked   bool    `json:"blocked,omitempty"`
}

// TokenLogProb represents token log probability
type TokenLogProb struct {
	Token       string  `json:"token"`
	LogProbability float64 `json:"logProbability"`
	TopCandidates []TopCandidate `json:"topCandidates,omitempty"`
}

// TopCandidate represents a top candidate
type TopCandidate struct {
	Token string `json:"token"`
	LogProbability float64 `json:"logProbability"`
}

// UsageMetadata represents usage metadata
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount     int `json:"totalTokenCount"`
	CachedContentTokenCount int `json:"cachedContentTokenCount,omitempty"`
}

// PromptFeedback represents prompt feedback
type PromptFeedback struct {
	BlockReason   string `json:"blockReason"` // "BLOCK_REASON_UNSPECIFIED", "SAFETY", "OTHER"
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// ErrorResponse represents Gemini API error response
type ErrorResponse struct {
	Error *Error `json:"error"`
}

// Error represents an error detail
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
	Details []Detail `json:"details,omitempty"`
}

// Detail represents error detail
type Detail struct {
	Type string `json:"@type"`
	Rest map[string]interface{} `json:"-"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Candidates      []Candidate   `json:"candidates"`
	UsageMetadata   *UsageMetadata `json:"usageMetadata,omitempty"`
	ModelVersion    string `json:"modelVersion,omitempty"`
	PromptFeedback  *PromptFeedback `json:"promptFeedback,omitempty"`
}

// Supported Gemini models
var SupportedModels = []string{
	"gemini-2.5-pro",
	"gemini-2.5-flash",
	"gemini-2.0-flash",
}

// Constants for finish reasons
const (
	FinishReasonUnspecified = "FINISH_REASON_UNSPECIFIED"
	FinishReasonStop       = "STOP"
	FinishReasonMaxTokens  = "MAX_TOKENS"
	FinishReasonSafety     = "SAFETY"
	FinishReasonRecitation = "RECITATION"
	FinishReasonOther      = "OTHER"
)

// Constants for safety categories
const (
	SafetyCategoryHarassment      = "HARM_CATEGORY_HARASSMENT"
	SafetyCategoryHateSpeech     = "HARM_CATEGORY_HATE_SPEECH"
	SafetyCategorySexuallyExplicit = "HARM_CATEGORY_SEXUALLY_EXPLICIT"
	SafetyCategoryDangerousContent = "HARM_CATEGORY_DANGEROUS_CONTENT"
)
