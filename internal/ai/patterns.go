package ai

// FrameworkInfo holds detection patterns for a single AI/ML framework.
type FrameworkInfo struct {
	Name     string
	Patterns []string // regex patterns for code imports
	DepFiles []string // files to check for dependencies
	DepNames []string // dependency names to search for
	Type     string   // FRAMEWORK, LLM_API, LLM_LOCAL, MLOPS
	RiskBase string   // HIGH, MEDIUM, LOW
}

// AIFrameworks is the complete registry of detectable AI/ML frameworks.
// Ported from Python AI_FRAMEWORKS in backend/core/ai_scanner.py
var AIFrameworks = map[string]FrameworkInfo{
	"tensorflow": {
		Name:     "TensorFlow",
		Patterns: []string{`import tensorflow`, `from tensorflow`, `tf\.keras`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml", "package.json"},
		DepNames: []string{"tensorflow", "tensorflow-gpu", "tf-nightly"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
	"pytorch": {
		Name:     "PyTorch",
		Patterns: []string{`import torch`, `from torch`, `torch\.nn`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"torch", "torchvision", "pytorch"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
	"scikit-learn": {
		Name:     "scikit-learn",
		Patterns: []string{`from sklearn`, `import sklearn`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"scikit-learn", "sklearn"},
		Type:     "FRAMEWORK",
		RiskBase: "MEDIUM",
	},
	"huggingface": {
		Name:     "HuggingFace",
		Patterns: []string{`from transformers`, `import transformers`, `AutoModel`, `AutoTokenizer`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"transformers", "huggingface-hub", "datasets"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
	"langchain": {
		Name:     "LangChain",
		Patterns: []string{`from langchain`, `import langchain`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"langchain", "langchain-core", "langchain-openai", "langchain-anthropic", "langchain-google"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
	// --- Enterprise LLM APIs ---
	"openai": {
		Name:     "OpenAI",
		Patterns: []string{`import openai`, `from openai`, `openai\.`, `OpenAI\(`, `ChatCompletion`, `OPENAI_API_KEY`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml", "package.json", ".env", ".env.example"},
		DepNames: []string{"openai"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"anthropic": {
		Name:     "Anthropic",
		Patterns: []string{`import anthropic`, `from anthropic`, `anthropic\.`, `Anthropic\(`, `ANTHROPIC_API_KEY`, `claude`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml", "package.json", ".env", ".env.example"},
		DepNames: []string{"anthropic"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"azure-openai": {
		Name:     "Azure OpenAI",
		Patterns: []string{`AzureOpenAI`, `azure\.ai\.openai`, `AZURE_OPENAI`, `openai\.api_type.*azure`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml", "package.json", ".env"},
		DepNames: []string{"openai", "azure-ai-openai"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"google-vertex-ai": {
		Name:     "Google Vertex AI",
		Patterns: []string{`vertexai`, `google\.cloud\.aiplatform`, `from google\.generativeai`, `import google\.generativeai`, `GOOGLE_API_KEY`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"google-cloud-aiplatform", "vertexai", "google-generativeai"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"aws-bedrock": {
		Name:     "AWS Bedrock",
		Patterns: []string{`bedrock`, `boto3.*bedrock`, `invoke_model`, `AWS_BEDROCK`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"boto3", "botocore"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"cohere": {
		Name:     "Cohere",
		Patterns: []string{`import cohere`, `from cohere`, `cohere\.`, `Cohere\(`, `COHERE_API_KEY`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml", "package.json"},
		DepNames: []string{"cohere"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"ai21": {
		Name:     "AI21 Labs",
		Patterns: []string{`import ai21`, `from ai21`, `ai21\.`, `AI21_API_KEY`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"ai21"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"mistral": {
		Name:     "Mistral AI",
		Patterns: []string{`import mistralai`, `from mistralai`, `MistralClient`, `MISTRAL_API_KEY`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"mistralai"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"replicate": {
		Name:     "Replicate",
		Patterns: []string{`import replicate`, `from replicate`, `replicate\.`, `REPLICATE_API_TOKEN`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"replicate"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"together-ai": {
		Name:     "Together AI",
		Patterns: []string{`import together`, `from together`, `together\.`, `TOGETHER_API_KEY`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"together"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"groq": {
		Name:     "Groq",
		Patterns: []string{`import groq`, `from groq`, `Groq\(`, `GROQ_API_KEY`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"groq"},
		Type:     "LLM_API",
		RiskBase: "HIGH",
	},
	"ollama": {
		Name:     "Ollama",
		Patterns: []string{`import ollama`, `from ollama`, `ollama\.`, `OLLAMA`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"ollama"},
		Type:     "LLM_LOCAL",
		RiskBase: "MEDIUM",
	},
	"llama-cpp": {
		Name:     "llama.cpp",
		Patterns: []string{`llama_cpp`, `from llama_cpp`, `LlamaCpp`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"llama-cpp-python"},
		Type:     "LLM_LOCAL",
		RiskBase: "MEDIUM",
	},
	"keras": {
		Name:     "Keras",
		Patterns: []string{`import keras`, `from keras`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"keras"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
	"xgboost": {
		Name:     "XGBoost",
		Patterns: []string{`import xgboost`, `from xgboost`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"xgboost"},
		Type:     "FRAMEWORK",
		RiskBase: "MEDIUM",
	},
	"lightgbm": {
		Name:     "LightGBM",
		Patterns: []string{`import lightgbm`, `from lightgbm`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"lightgbm"},
		Type:     "FRAMEWORK",
		RiskBase: "MEDIUM",
	},
	"mlflow": {
		Name:     "MLflow",
		Patterns: []string{`import mlflow`, `from mlflow`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"mlflow"},
		Type:     "MLOPS",
		RiskBase: "LOW",
	},
	"llamaindex": {
		Name:     "LlamaIndex",
		Patterns: []string{`from llama_index`, `import llama_index`, `LlamaIndex`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"llama-index", "llama_index"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
	"semantic-kernel": {
		Name:     "Semantic Kernel",
		Patterns: []string{`semantic_kernel`, `from semantic_kernel`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"semantic-kernel"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
	"autogen": {
		Name:     "AutoGen",
		Patterns: []string{`import autogen`, `from autogen`, `pyautogen`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"pyautogen", "autogen"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
	"crewai": {
		Name:     "CrewAI",
		Patterns: []string{`import crewai`, `from crewai`, `CrewAI`},
		DepFiles: []string{"requirements.txt", "setup.py", "pyproject.toml"},
		DepNames: []string{"crewai"},
		Type:     "FRAMEWORK",
		RiskBase: "HIGH",
	},
}

// ModelFileExtension maps a file extension to its model type description and risk level.
type ModelFileExtension struct {
	Extension string
	ModelType string
	Risk      string
}

// ModelFileExtensions lists all known AI model file extensions.
var ModelFileExtensions = []ModelFileExtension{
	{".h5", "Keras/TensorFlow Model", "HIGH"},
	{".hdf5", "HDF5 Model", "HIGH"},
	{".pt", "PyTorch Model", "HIGH"},
	{".pth", "PyTorch Model", "HIGH"},
	{".onnx", "ONNX Model", "HIGH"},
	{".pkl", "Pickled Model", "MEDIUM"},
	{".joblib", "Joblib Model", "MEDIUM"},
	{".safetensors", "SafeTensors Model", "HIGH"},
	{".bin", "Binary Model (potential)", "MEDIUM"},
	{".ckpt", "Checkpoint Model", "HIGH"},
	{".pb", "TensorFlow Protobuf", "HIGH"},
}

// AIConfigFiles are filenames that indicate AI configuration or training.
var AIConfigFiles = []string{
	"model_config.json",
	"config.json",
	"hyperparams.yaml",
	"hyperparameters.json",
	"training_config.yaml",
	"model.yaml",
	".mlflow",
	"mlproject",
	"dvc.yaml",
}
