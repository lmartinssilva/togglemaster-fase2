import re
import sys

def fix_file(path):
    with open(path, 'r', encoding='utf-8') as f:
        content = f.read()

    original = content

    # Só mexe em chaves que estão dentro de uma f-string (precedidas por f" ou f'
    # em algum ponto antes, na mesma "janela" de string) e que contêm quebra de linha.
    # Padrão: f"...{  <quebra de linha e espacos>  algo}..."
    pattern = re.compile(
        r'(f["\'][^"\'\n]*\{)([^{}]*?\n[^{}]*?)(\})',
        re.MULTILINE
    )

    def collapse(match):
        prefix = match.group(1)
        inner = match.group(2)
        suffix = match.group(3)
        collapsed_inner = re.sub(r'\s+', ' ', inner).strip()
        return prefix + collapsed_inner + suffix

    fixed = pattern.sub(collapse, content)

    if fixed != original:
        with open(path, 'w', encoding='utf-8') as f:
            f.write(fixed)
        print(f"Corrigido: {path}")
        # Mostra o diff simples de quais linhas mudaram
        import difflib
        diff = difflib.unified_diff(
            original.splitlines(keepends=True),
            fixed.splitlines(keepends=True),
            fromfile='antes', tofile='depois'
        )
        print(''.join(diff))
    else:
        print(f"Nenhuma mudança necessária: {path}")

for p in sys.argv[1:]:
    fix_file(p)
