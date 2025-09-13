const acorn = require('acorn');
const escodegen = require('escodegen');
const estraverse = require('estraverse');

// 变量名生成器
let varCounter = 0;
function generateVarName() {
    return 'var_' + varCounter++;
}

// 判断是否为恒定布尔表达式
function isAlwaysTrue(node) {
    if (node.type === 'UnaryExpression' && node.operator === '!' && node.argument.type === 'Literal') {
        return !node.argument.value;
    }
    if (node.type === 'Literal' && node.value === true) return true;
    if (node.type === 'BinaryExpression' && node.operator === '==' && node.left.type === 'Literal' && node.right.type === 'Literal') {
        return node.left.value == node.right.value;
    }
    if (node.type === 'BinaryExpression' && node.operator === '===' && node.left.type === 'Literal' && node.right.type === 'Literal') {
        return node.left.value === node.right.value;
    }
    return false;
}

function isAlwaysFalse(node) {
    if (node.type === 'UnaryExpression' && node.operator === '!' && node.argument.type === 'Literal') {
        return !!node.argument.value;
    }
    if (node.type === 'Literal' && node.value === false) return true;
    if (node.type === 'BinaryExpression' && node.operator === '==' && node.left.type === 'Literal' && node.right.type === 'Literal') {
        return node.left.value != node.right.value;
    }
    if (node.type === 'BinaryExpression' && node.operator === '===' && node.left.type === 'Literal' && node.right.type === 'Literal') {
        return node.left.value !== node.right.value;
    }
    return false;
}

// 检测是否为死循环
function isDeadLoop(node) {
    if (node.type === 'WhileStatement' && isAlwaysFalse(node.test)) return true;
    if (node.type === 'ForStatement') {
        if (node.test && isAlwaysFalse(node.test)) return true;
        if (node.init && node.init.type === 'VariableDeclaration') {
            const init = node.init.declarations[0];
            if (init && init.init && node.test && node.test.type === 'BinaryExpression') {
                // 检查 for (let i = 0; i < 0; i++)
                if (init.init.value === 0 && node.test.right.value === 0 &&
                    ['<', '<=', '=='].includes(node.test.operator)) {
                    return true;
                }
            }
        }
    }
    return false;
}

// 主去混淆函数
function deobfuscate(code) {
    try {
        const ast = acorn.parse(code, { ecmaVersion: 'latest', sourceType: 'script' });
        const varMap = new Map();
        const stringArrays = new Map(); // 存储字符串数组映射

        // 第一遍遍历：收集字符串数组
        estraverse.traverse(ast, {
            enter: (node) => {
                if (node.type === 'VariableDeclarator' && node.init && node.init.type === 'ArrayExpression') {
                    const arrayName = node.id.name;
                    const elements = node.init.elements.filter(el => el && el.type === 'Literal');
                    if (elements.length > 0 && elements.every(el => typeof el.value === 'string')) {
                        stringArrays.set(arrayName, elements.map(el => el.value));
                    }
                }
            }
        });

        // 第二遍遍历：执行去混淆
        estraverse.replace(ast, {
            enter: (node, parent) => {
                // 十六进制和Base64解码
                if (node.type === 'Literal' && typeof node.value === 'string') {
                    try {
                        if (node.value.match(/^0x[0-9a-fA-F]+$/)) {
                            const decoded = Buffer.from(node.value.slice(2), 'hex').toString('utf8');
                            return { type: 'Literal', value: decoded, raw: `'${decoded}'` };
                        }
                        if (node.value.match(/^[A-Za-z0-9+/=]+$/) && node.value.length > 0) {
                            try {
                                const decoded = Buffer.from(node.value, 'base64').toString('utf8');
                                if (decoded && decoded.length > 0 && !decoded.includes('\uFFFD')) {
                                    return { type: 'Literal', value: decoded, raw: `'${decoded}'` };
                                }
                            } catch (e) {}
                        }
                    } catch (e) {}
                }

                // 数组花指令还原 - 改进版本
                if (node.type === 'MemberExpression' && node.object.type === 'Identifier' &&
                    node.property.type === 'Literal' && typeof node.property.value === 'number') {
                    const arrayName = node.object.name;
                    if (stringArrays.has(arrayName)) {
                        const array = stringArrays.get(arrayName);
                        const index = node.property.value;
                        if (index >= 0 && index < array.length) {
                            return { type: 'Literal', value: array[index], raw: `'${array[index]}'` };
                        }
                    }
                }

                // 布尔表达式简化
                if (node.type === 'UnaryExpression' && node.operator === '!') {
                    if (node.argument.type === 'ArrayExpression' && node.argument.elements.length === 0) {
                        return { type: 'Literal', value: false, raw: 'false' };
                    }
                    if (node.argument.type === 'Literal') {
                        return { type: 'Literal', value: !node.argument.value, raw: (!node.argument.value).toString() };
                    }
                }
                if (node.type === 'UnaryExpression' && node.operator === '!' && node.argument.type === 'UnaryExpression' && node.argument.operator === '!') {
                    if (node.argument.argument.type === 'ArrayExpression' && node.argument.argument.elements.length === 0) {
                        return { type: 'Literal', value: true, raw: 'true' };
                    }
                    if (node.argument.argument.type === 'Literal') {
                        return { type: 'Literal', value: !!node.argument.argument.value, raw: (!!node.argument.argument.value).toString() };
                    }
                }

                // 死代码删除
                if ((node.type === 'IfStatement' || node.type === 'WhileStatement') && isAlwaysFalse(node.test)) {
                    return estraverse.VisitorOption.Remove;
                }
                if (node.type === 'IfStatement' && isAlwaysTrue(node.test)) {
                    return node.consequent;
                }
                if (isDeadLoop(node)) {
                    return estraverse.VisitorOption.Remove;
                }

                // 字符串合并
                if (node.type === 'BinaryExpression' && node.operator === '+' && node.left.type === 'Literal' && node.right.type === 'Literal') {
                    if (typeof node.left.value === 'string' && typeof node.right.value === 'string') {
                        const combined = node.left.value + node.right.value;
                        return { type: 'Literal', value: combined, raw: `'${combined}'` };
                    }
                }

                // 常量折叠
                if (node.type === 'BinaryExpression' && ['+', '-', '*', '/'].includes(node.operator)) {
                    if (node.left.type === 'Literal' && node.right.type === 'Literal' && typeof node.left.value === 'number' && typeof node.right.value === 'number') {
                        let value;
                        switch (node.operator) {
                            case '+': value = node.left.value + node.right.value; break;
                            case '-': value = node.left.value - node.right.value; break;
                            case '*': value = node.left.value * node.right.value; break;
                            case '/': value = node.left.value / node.right.value; break;
                        }
                        return { type: 'Literal', value, raw: `${value}` };
                    }
                }

                // 变量名重命名
                if (node.type === 'VariableDeclarator' && node.id.type === 'Identifier' && node.id.name.match(/^_0x[0-9a-f]+$/)) {
                    const newName = generateVarName();
                    varMap.set(node.id.name, newName);
                    node.id.name = newName;
                }
                if (node.type === 'Identifier' && varMap.has(node.name)) {
                    return { type: 'Identifier', name: varMap.get(node.name) };
                }

                // 自执行函数简化
                if (node.type === 'CallExpression' && node.callee.type === 'FunctionExpression' && node.arguments.length === 0) {
                    if (node.callee.body.body.length === 1 && node.callee.body.body[0].type === 'ReturnStatement') {
                        return node.callee.body.body[0].argument;
                    }
                }

                return node;
            }
        });

        try {
            return escodegen.generate(ast, {
                format: { indent: { style: '  ' }, quotes: 'single' }
            });
        } catch (e) {
            console.error('去混淆错误:', e.message);
            return code;
        }
    } catch (e) {
        console.error('去混淆错误:', e.message);
        return code;
    }
}

// 命令行参数支持
if (require.main === module) {
    const fs = require('fs');
    const path = process.argv[2];
    if (path) {
        const input = fs.readFileSync(path, 'utf-8');
        process.stdout.write(deobfuscate(input));
    } else {
        const obfuscatedCode = `
            var _0x1234 = ['a' + 'b', '0x48656c6c6f', ''];
            console.log(_0x1234[0], _0x1234[1]);
            ![];
            var _0x5678 = 1 + 2;
            (function() { return 'test'; })();
        `;
        console.log('原始代码:\n', obfuscatedCode);
        console.log('去混淆后代码:\n', deobfuscate(obfuscatedCode));
    }
}

module.exports = { deobfuscate };