package imports

type InvokeType string

const (
	INVOKENONSTATIC InvokeType = "invokenonstatic"
	INVOKESTATIC    InvokeType = "invokestatic"
	INVOKECLASS     InvokeType = "invokeclass"
	INVOKEABSTRACT  InvokeType = "invokeabstract"
	INVOKEVIRTUAL   InvokeType = "invokevirtual"
)

func getInvokeType(annotation string, str string) InvokeType {
	switch annotation {
	case "@staticmethod":
		return INVOKESTATIC
	case "@classmethod":
		return INVOKECLASS
	case "@abstractmethod":
		return INVOKEABSTRACT
	default:
		if str != "" {
			return INVOKEVIRTUAL
		}
		return INVOKENONSTATIC
	}
}
