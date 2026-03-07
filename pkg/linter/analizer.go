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
	Doc: `Проверяет сообщения логгеров (log, slog, zap и объекта logger) на соответствие правилам стиля проекта.

Основные возможности:
	1. Регистр: Сообщение должно начинаться со строчной буквы (игнорируя цифры и знаки);
	2. Язык: Сообщения должны быть на английском языке (латиница);
	3. Символы: Запрещено использование спецсимволов и эмодзи (не-ASCII символы);
	4. Безопасность: Поиск утечек чувствительных данных (password, token, api_key, secret) 
	   как в тексте сообщений, так и в именах передаваемых переменных.

Поддерживаемые возможности анализа:
	- Анализ констант (const) и строковых литералов;
	- Обработка конкатенации строк (напр. "error: " + reason);
	- Рекурсивный анализ внутри fmt.Sprintf;
	- Проверка всех аргументов функции логирования;
	- Проверка всех аргументов функции логирования;
	- Есть возможность конфигурации, подсказок, настройки вывода ошибок (скопом или по одной).

Ограничения анализа:
	- Линтер не может анализировать строки, возвращаемые функциями в рантайме
		log.Print(msgFromFunc())
	- Линтер не отслеживает переменные типа если они были объявлены в одном месте, а использованы в другом. Работает только с константами;
	- Не проверяются поля структур или элементы мап если они не являются константами;
	- Если в проекте используется собственная функция-обертка над логгером – линтер её пропустит.
`,
	Run: run,
}

// ArgumentAnalyzer функциональный тип для анализа кода
type ArgumentAnalyzer func(pass model.Reporter, arg ast.Expr, config *config.Config)

var configPath string

func init() {
	Analyzer.Flags.StringVar(&configPath, "config", "config.yml", "path to config file")
}

func run(pass *analysis.Pass) (interface{}, error) {
	cfg := config.LoadConfig(configPath)

	wrapper := model.PassWrapper{Pass: pass}

	var checkArg ArgumentAnalyzer
	if cfg.Output.ErrorsAggregate {
		checkArg = baseAnalizer.BaseAggregateAnalyzer

	} else {
		checkArg = baseAnalizer.BaseAnalyzer
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 || !util.IsLogCall(call) {
				return true
			}

			for _, arg := range call.Args {
				checkArg(wrapper, arg, cfg)
			}
			return true
		})
	}
	return nil, nil
}
