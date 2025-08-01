package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/danielmiessler/fabric/internal/chat"
	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/util"
	"github.com/jessevdk/go-flags"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// Flags create flags struct. the users flags go into this, this will be passed to the chat struct in cli
type Flags struct {
	Pattern                         string            `short:"p" long:"pattern" yaml:"pattern" description:"Choose a pattern from the available patterns" default:""`
	PatternVariables                map[string]string `short:"v" long:"variable" description:"Values for pattern variables, e.g. -v=#role:expert -v=#points:30"`
	Context                         string            `short:"C" long:"context" description:"Choose a context from the available contexts" default:""`
	Session                         string            `long:"session" description:"Choose a session from the available sessions"`
	Attachments                     []string          `short:"a" long:"attachment" description:"Attachment path or URL (e.g. for OpenAI image recognition messages)"`
	Setup                           bool              `short:"S" long:"setup" description:"Run setup for all reconfigurable parts of fabric"`
	Temperature                     float64           `short:"t" long:"temperature" yaml:"temperature" description:"Set temperature" default:"0.7"`
	TopP                            float64           `short:"T" long:"topp" yaml:"topp" description:"Set top P" default:"0.9"`
	Stream                          bool              `short:"s" long:"stream" yaml:"stream" description:"Stream"`
	PresencePenalty                 float64           `short:"P" long:"presencepenalty" yaml:"presencepenalty" description:"Set presence penalty" default:"0.0"`
	Raw                             bool              `short:"r" long:"raw" yaml:"raw" description:"Use the defaults of the model without sending chat options (like temperature etc.) and use the user role instead of the system role for patterns."`
	FrequencyPenalty                float64           `short:"F" long:"frequencypenalty" yaml:"frequencypenalty" description:"Set frequency penalty" default:"0.0"`
	ListPatterns                    bool              `short:"l" long:"listpatterns" description:"List all patterns"`
	ListAllModels                   bool              `short:"L" long:"listmodels" description:"List all available models"`
	ListAllContexts                 bool              `short:"x" long:"listcontexts" description:"List all contexts"`
	ListAllSessions                 bool              `short:"X" long:"listsessions" description:"List all sessions"`
	UpdatePatterns                  bool              `short:"U" long:"updatepatterns" description:"Update patterns"`
	Message                         string            `hidden:"true" description:"Messages to send to chat"`
	Copy                            bool              `short:"c" long:"copy" description:"Copy to clipboard"`
	Model                           string            `short:"m" long:"model" yaml:"model" description:"Choose model"`
	ModelContextLength              int               `long:"modelContextLength" yaml:"modelContextLength" description:"Model context length (only affects ollama)"`
	Output                          string            `short:"o" long:"output" description:"Output to file" default:""`
	OutputSession                   bool              `long:"output-session" description:"Output the entire session (also a temporary one) to the output file"`
	LatestPatterns                  string            `short:"n" long:"latest" description:"Number of latest patterns to list" default:"0"`
	ChangeDefaultModel              bool              `short:"d" long:"changeDefaultModel" description:"Change default model"`
	YouTube                         string            `short:"y" long:"youtube" description:"YouTube video or play list \"URL\" to grab transcript, comments from it and send to chat or print it put to the console and store it in the output file"`
	YouTubePlaylist                 bool              `long:"playlist" description:"Prefer playlist over video if both ids are present in the URL"`
	YouTubeTranscript               bool              `long:"transcript" description:"Grab transcript from YouTube video and send to chat (it is used per default)."`
	YouTubeTranscriptWithTimestamps bool              `long:"transcript-with-timestamps" description:"Grab transcript from YouTube video with timestamps and send to chat"`
	YouTubeComments                 bool              `long:"comments" description:"Grab comments from YouTube video and send to chat"`
	YouTubeMetadata                 bool              `long:"metadata" description:"Output video metadata"`
	Language                        string            `short:"g" long:"language" description:"Specify the Language Code for the chat, e.g. -g=en -g=zh" default:""`
	ScrapeURL                       string            `short:"u" long:"scrape_url" description:"Scrape website URL to markdown using Jina AI"`
	ScrapeQuestion                  string            `short:"q" long:"scrape_question" description:"Search question using Jina AI"`
	Seed                            int               `short:"e" long:"seed" yaml:"seed" description:"Seed to be used for LMM generation"`
	WipeContext                     string            `short:"w" long:"wipecontext" description:"Wipe context"`
	WipeSession                     string            `short:"W" long:"wipesession" description:"Wipe session"`
	PrintContext                    string            `long:"printcontext" description:"Print context"`
	PrintSession                    string            `long:"printsession" description:"Print session"`
	HtmlReadability                 bool              `long:"readability" description:"Convert HTML input into a clean, readable view"`
	InputHasVars                    bool              `long:"input-has-vars" description:"Apply variables to user input"`
	DryRun                          bool              `long:"dry-run" description:"Show what would be sent to the model without actually sending it"`
	Serve                           bool              `long:"serve" description:"Serve the Fabric Rest API"`
	ServeOllama                     bool              `long:"serveOllama" description:"Serve the Fabric Rest API with ollama endpoints"`
	ServeAddress                    string            `long:"address" description:"The address to bind the REST API" default:":8080"`
	ServeAPIKey                     string            `long:"api-key" description:"API key used to secure server routes" default:""`
	Config                          string            `long:"config" description:"Path to YAML config file"`
	Version                         bool              `long:"version" description:"Print current version"`
	ListExtensions                  bool              `long:"listextensions" description:"List all registered extensions"`
	AddExtension                    string            `long:"addextension" description:"Register a new extension from config file path"`
	RemoveExtension                 string            `long:"rmextension" description:"Remove a registered extension by name"`
	Strategy                        string            `long:"strategy" description:"Choose a strategy from the available strategies" default:""`
	ListStrategies                  bool              `long:"liststrategies" description:"List all strategies"`
	ListVendors                     bool              `long:"listvendors" description:"List all vendors"`
	ShellCompleteOutput             bool              `long:"shell-complete-list" description:"Output raw list without headers/formatting (for shell completion)"`
	Search                          bool              `long:"search" description:"Enable web search tool for supported models (Anthropic, OpenAI)"`
	SearchLocation                  string            `long:"search-location" description:"Set location for web search results (e.g., 'America/Los_Angeles')"`
	ImageFile                       string            `long:"image-file" description:"Save generated image to specified file path (e.g., 'output.png')"`
	ImageSize                       string            `long:"image-size" description:"Image dimensions: 1024x1024, 1536x1024, 1024x1536, auto (default: auto)"`
	ImageQuality                    string            `long:"image-quality" description:"Image quality: low, medium, high, auto (default: auto)"`
	ImageCompression                int               `long:"image-compression" description:"Compression level 0-100 for JPEG/WebP formats (default: not set)"`
	ImageBackground                 string            `long:"image-background" description:"Background type: opaque, transparent (default: opaque, only for PNG/WebP)"`
	SuppressThink                   bool              `long:"suppress-think" yaml:"suppressThink" description:"Suppress text enclosed in thinking tags"`
	ThinkStartTag                   string            `long:"think-start-tag" yaml:"thinkStartTag" description:"Start tag for thinking sections" default:"<think>"`
	ThinkEndTag                     string            `long:"think-end-tag" yaml:"thinkEndTag" description:"End tag for thinking sections" default:"</think>"`
	DisableResponsesAPI             bool              `long:"disable-responses-api" yaml:"disableResponsesAPI" description:"Disable OpenAI Responses API (default: false)"`
	Voice                           string            `long:"voice" yaml:"voice" description:"TTS voice name for supported models (e.g., Kore, Charon, Puck)" default:"Kore"`
	ListGeminiVoices                bool              `long:"list-gemini-voices" description:"List all available Gemini TTS voices"`
}

var debug = false

func Debugf(format string, a ...interface{}) {
	if debug {
		fmt.Printf("DEBUG: "+format, a...)
	}
}

// Init Initialize flags. returns a Flags struct and an error
func Init() (ret *Flags, err error) {
	// Track which yaml-configured flags were set on CLI
	usedFlags := make(map[string]bool)
	yamlArgsScan := os.Args[1:]

	// Create mapping from flag names (both short and long) to yaml tag names
	flagToYamlTag := make(map[string]string)
	t := reflect.TypeOf(Flags{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag != "" {
			longTag := field.Tag.Get("long")
			shortTag := field.Tag.Get("short")
			if longTag != "" {
				flagToYamlTag[longTag] = yamlTag
				Debugf("Mapped long flag %s to yaml tag %s\n", longTag, yamlTag)
			}
			if shortTag != "" {
				flagToYamlTag[shortTag] = yamlTag
				Debugf("Mapped short flag %s to yaml tag %s\n", shortTag, yamlTag)
			}
		}
	}

	// Scan args for that are provided by cli and might be in yaml
	for _, arg := range yamlArgsScan {
		flag := extractFlag(arg)

		if flag != "" {
			if yamlTag, exists := flagToYamlTag[flag]; exists {
				usedFlags[yamlTag] = true
				Debugf("CLI flag used: %s (yaml: %s)\n", flag, yamlTag)
			}
		}
	}

	// Parse CLI flags first
	ret = &Flags{}
	parser := flags.NewParser(ret, flags.Default)
	var args []string
	if args, err = parser.Parse(); err != nil {
		return
	}

	// Check to see if a ~/.config/fabric/config.yaml config file exists (only when user didn't specify a config)
	if ret.Config == "" {
		// Default to ~/.config/fabric/config.yaml if no config specified
		if defaultConfigPath, err := util.GetDefaultConfigPath(); err == nil && defaultConfigPath != "" {
			ret.Config = defaultConfigPath
		} else if err != nil {
			Debugf("Could not determine default config path: %v\n", err)
		}
	}

	// If config specified, load and apply YAML for unused flags
	if ret.Config != "" {
		var yamlFlags *Flags
		if yamlFlags, err = loadYAMLConfig(ret.Config); err != nil {
			return
		}

		// Apply YAML values where CLI flags weren't used
		flagsVal := reflect.ValueOf(ret).Elem()
		yamlVal := reflect.ValueOf(yamlFlags).Elem()
		flagsType := flagsVal.Type()

		for i := 0; i < flagsType.NumField(); i++ {
			field := flagsType.Field(i)
			if yamlTag := field.Tag.Get("yaml"); yamlTag != "" {
				if !usedFlags[yamlTag] {
					flagField := flagsVal.Field(i)
					yamlField := yamlVal.Field(i)
					if flagField.CanSet() {
						if yamlField.Type() != flagField.Type() {
							if err := assignWithConversion(flagField, yamlField); err != nil {
								Debugf("Type conversion failed for %s: %v\n", yamlTag, err)
								continue
							}
						} else {
							flagField.Set(yamlField)
						}
						Debugf("Applied YAML value for %s: %v\n", yamlTag, yamlField.Interface())
					}
				}
			}
		}
	}

	// Handle stdin and messages
	info, _ := os.Stdin.Stat()
	pipedToStdin := (info.Mode() & os.ModeCharDevice) == 0

	// Append positional arguments to the message (custom message)
	if len(args) > 0 {
		ret.Message = AppendMessage(ret.Message, args[len(args)-1])
	}

	if pipedToStdin {
		var pipedMessage string
		if pipedMessage, err = readStdin(); err != nil {
			return
		}
		ret.Message = AppendMessage(ret.Message, pipedMessage)
	}
	return
}

func extractFlag(arg string) string {
	var flag string
	if strings.HasPrefix(arg, "--") {
		flag = strings.TrimPrefix(arg, "--")
		if i := strings.Index(flag, "="); i > 0 {
			flag = flag[:i]
		}
	} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
		flag = strings.TrimPrefix(arg, "-")
		if i := strings.Index(flag, "="); i > 0 {
			flag = flag[:i]
		}
	}
	return flag
}

func assignWithConversion(targetField, sourceField reflect.Value) error {
	// Handle string source values
	if sourceField.Kind() == reflect.String {
		str := sourceField.String()
		switch targetField.Kind() {
		case reflect.Int:
			// Try parsing as float first to handle "42.9" -> 42
			if val, err := strconv.ParseFloat(str, 64); err == nil {
				targetField.SetInt(int64(val))
				return nil
			}
			// Try direct int parse
			if val, err := strconv.ParseInt(str, 10, 64); err == nil {
				targetField.SetInt(val)
				return nil
			}
		case reflect.Float64:
			if val, err := strconv.ParseFloat(str, 64); err == nil {
				targetField.SetFloat(val)
				return nil
			}
		case reflect.Bool:
			if val, err := strconv.ParseBool(str); err == nil {
				targetField.SetBool(val)
				return nil
			}
		}
		return fmt.Errorf("cannot convert string %q to %v", str, targetField.Kind())
	}

	return fmt.Errorf("unsupported conversion from %v to %v", sourceField.Kind(), targetField.Kind())
}

func loadYAMLConfig(configPath string) (*Flags, error) {
	absPath, err := util.GetAbsolutePath(configPath)
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", absPath)
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Use the existing Flags struct for YAML unmarshal
	config := &Flags{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	Debugf("Config: %v\n", config)

	return config, nil
}

// readStdin reads from stdin and returns the input as a string or an error
func readStdin() (ret string, err error) {
	reader := bufio.NewReader(os.Stdin)
	var sb strings.Builder
	for {
		if line, readErr := reader.ReadString('\n'); readErr != nil {
			if errors.Is(readErr, io.EOF) {
				sb.WriteString(line)
				break
			}
			err = fmt.Errorf("error reading piped message from stdin: %w", readErr)
			return
		} else {
			sb.WriteString(line)
		}
	}
	ret = sb.String()
	return
}

// validateImageFile validates the image file path and extension
func validateImageFile(imagePath string) error {
	if imagePath == "" {
		return nil // No validation needed if no image file specified
	}

	// Check if file already exists
	if _, err := os.Stat(imagePath); err == nil {
		return fmt.Errorf("image file already exists: %s", imagePath)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(imagePath))
	validExtensions := []string{".png", ".jpeg", ".jpg", ".webp"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return nil // Valid extension found
		}
	}

	return fmt.Errorf("invalid image file extension '%s'. Supported formats: .png, .jpeg, .jpg, .webp", ext)
}

// validateImageParameters validates image generation parameters
func validateImageParameters(imagePath, size, quality, background string, compression int) error {
	if imagePath == "" {
		// Check if any image parameters are specified without --image-file
		if size != "" || quality != "" || background != "" || compression != 0 {
			return fmt.Errorf("image parameters (--image-size, --image-quality, --image-background, --image-compression) can only be used with --image-file")
		}
		return nil
	}

	// Validate size
	if size != "" {
		validSizes := []string{"1024x1024", "1536x1024", "1024x1536", "auto"}
		valid := false
		for _, validSize := range validSizes {
			if size == validSize {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid image size '%s'. Supported sizes: 1024x1024, 1536x1024, 1024x1536, auto", size)
		}
	}

	// Validate quality
	if quality != "" {
		validQualities := []string{"low", "medium", "high", "auto"}
		valid := false
		for _, validQuality := range validQualities {
			if quality == validQuality {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid image quality '%s'. Supported qualities: low, medium, high, auto", quality)
		}
	}

	// Validate background
	if background != "" {
		validBackgrounds := []string{"opaque", "transparent"}
		valid := false
		for _, validBackground := range validBackgrounds {
			if background == validBackground {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid image background '%s'. Supported backgrounds: opaque, transparent", background)
		}
	}

	// Get file format for format-specific validations
	ext := strings.ToLower(filepath.Ext(imagePath))

	// Validate compression (only for jpeg/webp)
	if compression != 0 { // 0 means not set
		if ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
			return fmt.Errorf("image compression can only be used with JPEG and WebP formats, not %s", ext)
		}
		if compression < 0 || compression > 100 {
			return fmt.Errorf("image compression must be between 0 and 100, got %d", compression)
		}
	}

	// Validate background transparency (only for png/webp)
	if background == "transparent" {
		if ext != ".png" && ext != ".webp" {
			return fmt.Errorf("transparent background can only be used with PNG and WebP formats, not %s", ext)
		}
	}

	return nil
}

func (o *Flags) BuildChatOptions() (ret *domain.ChatOptions, err error) {
	// Validate image file if specified
	if err = validateImageFile(o.ImageFile); err != nil {
		return nil, err
	}

	// Validate image parameters
	if err = validateImageParameters(o.ImageFile, o.ImageSize, o.ImageQuality, o.ImageBackground, o.ImageCompression); err != nil {
		return nil, err
	}

	startTag := o.ThinkStartTag
	if startTag == "" {
		startTag = "<think>"
	}
	endTag := o.ThinkEndTag
	if endTag == "" {
		endTag = "</think>"
	}

	ret = &domain.ChatOptions{
		Model:              o.Model,
		Temperature:        o.Temperature,
		TopP:               o.TopP,
		PresencePenalty:    o.PresencePenalty,
		FrequencyPenalty:   o.FrequencyPenalty,
		Raw:                o.Raw,
		Seed:               o.Seed,
		ModelContextLength: o.ModelContextLength,
		Search:             o.Search,
		SearchLocation:     o.SearchLocation,
		ImageFile:          o.ImageFile,
		ImageSize:          o.ImageSize,
		ImageQuality:       o.ImageQuality,
		ImageCompression:   o.ImageCompression,
		ImageBackground:    o.ImageBackground,
		SuppressThink:      o.SuppressThink,
		ThinkStartTag:      startTag,
		ThinkEndTag:        endTag,
		Voice:              o.Voice,
	}
	return
}

func (o *Flags) BuildChatRequest(Meta string) (ret *domain.ChatRequest, err error) {
	ret = &domain.ChatRequest{
		ContextName:      o.Context,
		SessionName:      o.Session,
		PatternName:      o.Pattern,
		StrategyName:     o.Strategy,
		PatternVariables: o.PatternVariables,
		InputHasVars:     o.InputHasVars,
		Meta:             Meta,
	}

	var message *chat.ChatCompletionMessage
	if len(o.Attachments) > 0 {
		message = &chat.ChatCompletionMessage{
			Role: chat.ChatMessageRoleUser,
		}

		if o.Message != "" {
			message.MultiContent = append(message.MultiContent, chat.ChatMessagePart{
				Type: chat.ChatMessagePartTypeText,
				Text: strings.TrimSpace(o.Message),
			})
		}

		for _, attachmentValue := range o.Attachments {
			var attachment *domain.Attachment
			if attachment, err = domain.NewAttachment(attachmentValue); err != nil {
				return
			}
			url := attachment.URL
			if url == nil {
				var base64Image string
				if base64Image, err = attachment.Base64Content(); err != nil {
					return
				}
				var mimeType string
				if mimeType, err = attachment.ResolveType(); err != nil {
					return
				}
				dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)
				url = &dataURL
			}
			message.MultiContent = append(message.MultiContent, chat.ChatMessagePart{
				Type: chat.ChatMessagePartTypeImageURL,
				ImageURL: &chat.ChatMessageImageURL{
					URL: *url,
				},
			})
		}
	} else if o.Message != "" {
		message = &chat.ChatCompletionMessage{
			Role:    chat.ChatMessageRoleUser,
			Content: strings.TrimSpace(o.Message),
		}
	}

	ret.Message = message

	if o.Language != "" {
		if langTag, langErr := language.Parse(o.Language); langErr == nil {
			ret.Language = langTag.String()
		}
	}
	return
}

func (o *Flags) AppendMessage(message string) {
	o.Message = AppendMessage(o.Message, message)
}

func (o *Flags) IsChatRequest() (ret bool) {
	ret = o.Message != "" || len(o.Attachments) > 0 || o.Context != "" || o.Session != "" || o.Pattern != ""
	return
}

func (o *Flags) WriteOutput(message string) (err error) {
	fmt.Println(message)
	if o.Output != "" {
		err = CreateOutputFile(message, o.Output)
	}
	return
}

func AppendMessage(message string, newMessage string) (ret string) {
	if message != "" {
		ret = message + "\n" + newMessage
	} else {
		ret = newMessage
	}
	return
}
