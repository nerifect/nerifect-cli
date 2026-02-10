package ai

// classifyEUAIActRisk maps AI component types to EU AI Act risk levels.
func classifyEUAIActRisk(componentType string) string {
	switch componentType {
	case "FRAMEWORK":
		return "HIGH-RISK"
	case "LLM_API":
		return "LIMITED-RISK"
	case "LLM_LOCAL":
		return "LIMITED-RISK"
	case "MLOPS":
		return "MINIMAL-RISK"
	case "CONFIG":
		return "MINIMAL-RISK"
	default:
		// Model files default to HIGH-RISK since they may involve
		// critical infrastructure or decision-making
		if componentType == "Keras/TensorFlow Model" ||
			componentType == "PyTorch Model" ||
			componentType == "ONNX Model" ||
			componentType == "SafeTensors Model" ||
			componentType == "Checkpoint Model" ||
			componentType == "TensorFlow Protobuf" {
			return "HIGH-RISK"
		}
		return "LIMITED-RISK"
	}
}
