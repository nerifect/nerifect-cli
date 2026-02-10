# AI/ML Detection

Nerifect detects AI and ML framework usage in your codebase using a 4-phase detection engine. Detected components are classified under the EU AI Act risk framework.

## Supported Frameworks

### ML Frameworks

| Framework | Detection Methods |
|---|---|
| TensorFlow | Import patterns, `requirements.txt`, model files (`.h5`, `.pb`) |
| PyTorch | Import patterns, `requirements.txt`, model files (`.pt`, `.pth`) |
| scikit-learn | Import patterns, `requirements.txt`, model files (`.pkl`) |
| Keras | Import patterns, `requirements.txt`, model files (`.keras`) |
| XGBoost | Import patterns, `requirements.txt` |
| LightGBM | Import patterns, `requirements.txt` |

### LLM API Providers

| Provider | Detection Methods |
|---|---|
| OpenAI | Import patterns, `package.json`, config files |
| Anthropic | Import patterns, `package.json`, config files |
| Google Vertex AI | Import patterns, config files |
| Azure OpenAI | Import patterns, config files |
| AWS Bedrock | Import patterns, config files |
| Cohere | Import patterns, `requirements.txt` |
| AI21 | Import patterns, `requirements.txt` |
| Mistral | Import patterns, config files |
| Replicate | Import patterns, config files |
| Together AI | Import patterns, config files |
| Groq | Import patterns, config files |

### Local LLM Runtimes

| Runtime | Detection Methods |
|---|---|
| Ollama | Config files, import patterns |
| llama.cpp | Model files (`.gguf`), config files |

### Orchestration Frameworks

| Framework | Detection Methods |
|---|---|
| LangChain | Import patterns, `requirements.txt`, `package.json` |
| LlamaIndex | Import patterns, `requirements.txt` |
| Semantic Kernel | Import patterns, `package.json` |
| AutoGen | Import patterns, `requirements.txt` |
| CrewAI | Import patterns, `requirements.txt` |

### Pre-trained Models

| Framework | Detection Methods |
|---|---|
| HuggingFace Transformers | Import patterns, `requirements.txt`, model files (`.safetensors`) |

### MLOps

| Framework | Detection Methods |
|---|---|
| MLflow | Import patterns, `requirements.txt`, config files |

## Detection Phases

Nerifect runs detection in 4 phases, ordered by specificity:

### Phase 1: Model Files

Scans for files with known model extensions:

- `.pt`, `.pth` --- PyTorch models
- `.h5`, `.keras` --- TensorFlow/Keras models
- `.onnx` --- ONNX models
- `.safetensors` --- HuggingFace models
- `.pb` --- TensorFlow Protocol Buffer models
- `.pkl` --- scikit-learn models
- `.gguf` --- llama.cpp models
- `.tflite` --- TensorFlow Lite models

### Phase 2: Config Files

Looks for AI-specific configuration files such as `openai.yaml`, `.env` files with AI API keys, `mlflow` configs, etc.

### Phase 3: Dependency Files

Checks package managers for AI library dependencies:

- `requirements.txt` (Python)
- `package.json` (Node.js)
- `go.mod` (Go)
- `Gemfile` (Ruby)
- `pom.xml` (Java)

### Phase 4: Code Imports

Pattern-matches import statements in source files:

- Python: `import openai`, `from langchain import ...`
- JavaScript: `require('openai')`, `import { OpenAI } from 'openai'`
- Go: `import "github.com/sashabaranov/go-openai"`

## EU AI Act Risk Classification

Detected AI components are automatically classified under the EU AI Act:

| Risk Level | Description | Examples |
|---|---|---|
| **HIGH-RISK** | AI systems in critical domains requiring strict compliance | Biometric identification, critical infrastructure, employment decisions |
| **LIMITED-RISK** | AI systems with transparency obligations | Chatbots, emotion recognition, deepfake generation |
| **MINIMAL-RISK** | AI systems with minimal regulatory requirements | Spam filters, content recommendation, game AI |

Risk classification is performed using a combination of:

1. **Rule-based heuristics** --- Component type and usage patterns
2. **LLM assessment** --- The configured LLM provider analyzes detected components in context to refine risk levels

!!! note
    AI detection works without an API key (phases 1-4 are purely pattern-based). LLM risk assessment requires a valid API key for your configured provider.
