package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// 定义操作符结构体
type Operator struct {
	Func       func(int, int) int
	Precedence int
	Name       string
}

// 定义表达式树节点结构体
type TreeNode struct {
	Type  int // 0: 操作数, 1: 操作符
	Data  interface{}
	Left  *TreeNode
	Right *TreeNode
}

// 操作符函数定义
func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func mul(a, b int) int {
	return a * b
}

func div(a, b int) int {
	if b != 0 {
		return a / b
	}
	fmt.Println("除数不能为0")
	return 0 // 这里不再直接退出程序，而是返回0表示错误结果
}

// 预定义的操作符数组
var operators = []Operator{
	{add, 1, "add"},
	{sub, 1, "sub"},
	{mul, 2, "mul"},
	{div, 2, "div"},
	{add, 1, "+"},
	{sub, 1, "-"},
	{mul, 2, "*"},
	{div, 2, "/"},
}

// 根据操作符字符串查找对应的函数指针和优先级
func findOperator(opStr string) Operator {
	for _, op := range operators {
		if op.Name == opStr {
			return op
		}
	}
	fmt.Printf("无效操作符: %s\n", opStr)
	return Operator{nil, 0, "unknown"}
}

// 解析操作数
func parseOperand(input *string) (*TreeNode, bool) {
	trimmed := strings.TrimLeft(*input, " \t")
	var num int
	var err error
	var sign = 1

	// 处理负数
	if strings.HasPrefix(trimmed, "-") {
		sign = -1
		trimmed = trimmed[1:] // 移除负号
	}

	// 解析数字
	numStr := ""
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] >= '0' && trimmed[i] <= '9' {
			numStr += string(trimmed[i])
		} else {
			break
		}
	}
	if numStr == "" {
		return nil, false
	}
	num, err = strconv.Atoi(numStr)
	if err != nil {
		return nil, false
	}
	num *= sign

	// 更新输入字符串，去掉已解析的部分
	*input = strings.TrimLeft(trimmed[len(numStr):], " \t")

	fmt.Printf("num: %d\n", num)
	return &TreeNode{
		Type:  0,
		Data:  num,
		Left:  nil,
		Right: nil,
	}, true
}

// 解析操作符
func parseOperator(input *string) Operator {
	trimmed := strings.TrimLeft(*input, " \t")
	// 检查第一个字符是否为操作符符号
	for _, op := range operators {
		if len(trimmed) > 0 && strings.HasPrefix(trimmed, op.Name) {
			*input = strings.TrimLeft(trimmed[len(op.Name):], " \t")
			fmt.Printf("parsed operator: %s\n", op.Name)
			return op
		}
	}
	// 如果不是符号操作符，再按原逻辑匹配字母操作符
	opStr := ""
	for _, char := range trimmed {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
			opStr += string(char)
		} else {
			break
		}
	}
	if opStr == "" {
		if len(trimmed) > 0 && (trimmed[0] == '(' || trimmed[0] == ')') {
			fmt.Println("括号操作符")
		} else {
			fmt.Printf("无效字符，非操作符: '%c'\n", trimmed[0])
		}
		return Operator{nil, 0, "unknown"}
	}
	*input = strings.TrimLeft(trimmed[len(opStr):], " \t")
	op := findOperator(opStr)
	fmt.Printf("parsed operator: %s\n", opStr)
	return op
}

// 解析因子（数字或括号内的表达式）
func parseFactor(input *string) (*TreeNode, bool) {
	trimmed := strings.TrimLeft(*input, " \t")
	if len(trimmed) > 0 && trimmed[0] == '(' {
		fmt.Println("检测到左括号")
		*input = strings.TrimLeft(trimmed[1:], " \t")
		node, ok := parseExpression(input, 0)
		if !ok {
			return nil, false
		}
		trimmed = strings.TrimLeft(*input, " \t")
		if len(trimmed) > 0 && trimmed[0] == ')' {
			fmt.Println("检测到右括号")
			*input = strings.TrimLeft(trimmed[1:], " \t")
			return node, true
		} else {
			fmt.Println("缺少右括号")
			return nil, false
		}
	} else if len(trimmed) > 0 && (trimmed[0] >= '0' && trimmed[0] <= '9' || trimmed[0] == '-') {
		return parseOperand(input)
	} else {
		fmt.Printf("无效内容: '%c'\n", trimmed[0])
		return nil, false
	}
}

// 解析表达式（根据优先级）
func parseExpression(input *string, precedence int) (*TreeNode, bool) {
	left, ok := parseFactor(input)
	if !ok {
		return nil, false
	}
	for {
		trimmed := strings.TrimLeft(*input, " \t")
		if len(trimmed) == 0 {
			break
		}
		savedInput := *input
		op := parseOperator(input)
		if op.Precedence == 0 {
			*input = savedInput
			break
		}
		if op.Precedence <= precedence {
			*input = savedInput
			break
		}
		right, ok := parseExpression(input, op.Precedence)
		if !ok {
			return nil, false
		}
		newNode := &TreeNode{
			Type:  1,
			Data:  op,
			Left:  left,
			Right: right,
		}
		fmt.Printf("op: %s\n", op.Name)
		left = newNode
	}
	return left, true
}

// 构建表达式树
func buildExpressionTree(input string) (*TreeNode, bool, int) {
	root, ok := parseExpression(&input, 0)
	errorFlag := 0
	if !ok {
		errorFlag = 1
	}
	return root, ok, errorFlag
}

// 计算表达式树的结果
func evaluateExpressionTree(node *TreeNode) (int, int) {
	errorFlag := 0
	if node == nil {
		return 0, errorFlag
	}
	if node.Type == 0 {
		return node.Data.(int), errorFlag
	}
	op := node.Data.(Operator)
	leftValue, _ := evaluateExpressionTree(node.Left)
	rightValue, _ := evaluateExpressionTree(node.Right)
	result := op.Func(leftValue, rightValue)
	if op.Name == "div" && result == 0 { // 除数为0的情况
		errorFlag = 1
	}
	return result, errorFlag
}

// 定义请求结构体
type RequestData struct {
	Input string `json:"input"`
}

// 定义响应结构体
type ResponseData struct {
	Result    int `json:"result"`
	ErrorFlag int `json:"errorFlag"`
}

// 处理请求的函数
func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var request RequestData
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 构建表达式树并计算结果
	tree, ok, errorFlag := buildExpressionTree(request.Input)
	if !ok {
		errorFlag = 1
	}
	result, flag := evaluateExpressionTree(tree)
	fmt.Println("result =", result)
	for i := 0; i < 100; i++ {
		fmt.Print("=")
	}
	fmt.Println()

	errorFlag = errorFlag | flag // 合并错误标志

	// 构建响应
	response := ResponseData{
		Result:    result,
		ErrorFlag: errorFlag,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}
func logo() {
	fmt.Println("(´・ω・)つweb服务已启动，监听在 :8080 端口")
	for i := 0; i < 100; i++ {
		fmt.Print("-")
	}
	fmt.Println()
}
func main() {
	logo()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/api/process", handleRequest)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
