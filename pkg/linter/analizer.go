package linter

import (
	"github.com/MishaSytch/linter/pkg/config"
	"github.com/MishaSytch/linter/pkg/linter/model"
	"github.com/MishaSytch/linter/pkg/linter/util"
	"github.com/MishaSytch/linter/pkg/linter/util/baseAnalizer"
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// Analyzer определяет конфигурацию линтера для использования в основной программе.
var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc: `LogLint — это специализированный линтер для Go, разработанный для контроля качества логирования. 
Он проверяет сообщения логгеров (log, slog, zap и объекта logger) на соответствие правилам стиля проекта.

## Возможности

### 1. Основные правила проверки:
- **Регистр**: Сообщение должно начинаться со строчной буквы (игнорируя цифры и знаки)
- **Язык**: Сообщения должны быть на английском языке (латиница)
- **Символы**: Запрещено использование спецсимволов и эмодзи (не-ASCII символы)
- **Безопасность**: Поиск утечек чувствительных данных (password, token, api_key, secret) как в тексте сообщений, 
так и в именах передаваемых переменных

### 2. Поддерживаемые возможности анализа:
- Анализ констант (const) и строковых литералов
- Обработка конкатенации строк (например, "error: " + reason)
- Рекурсивный анализ внутри fmt.Sprintf
- Проверка всех аргументов функции логирования
- Конфигурация правил, вывод подсказок, настройка формата вывода ошибок (скопом или по одной)

### 3. Ограничения анализа:
- Линтер не может анализировать строки, возвращаемые функциями в рантайме (например, log.Print(msgFromFunc()))
- Если в проекте используется собственная функция-обертка над логгером Пример: 

	logErr := func(m string) {
		logger.Error(m)
	}
	logErr("Upper case error")
`,
	Run: run,
}

var configPath string

func init() {
	Analyzer.Flags.StringVar(&configPath, "config", "config.yml", "path to config file")
}

func run(pass *analysis.Pass) (interface{}, error) {
	cfg := config.LoadConfig(configPath)

	wrapper := model.PassWrapper{Pass: pass}

	var checkArg model.FunctionalAnalyzer
	if cfg.Output.ErrorsAggregate {
		checkArg = &baseAnalizer.BaseAggregateAnalyzer{
			Pass:   wrapper,
			Config: cfg,
		}

	} else {
		checkArg = &baseAnalizer.BaseAnalyzer{
			Pass:   wrapper,
			Config: cfg,
		}
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 || !util.IsLogCall(wrapper, call) {
				return true
			}

			for _, arg := range call.Args {
				checkArg.Analyze(arg)
			}
			return true
		})
	}
	return nil, nil
}
