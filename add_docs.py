import os, re

def add_go_docs(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()
    
    out = []
    i = 0
    while i < len(lines):
        line = lines[i]
        match = re.match(r'^func\s+([A-Za-z0-9_]+)\(', line)
        if match:
            func_name = match.group(1)
            if i == 0 or not lines[i-1].strip().startswith('//'):
                out.append(f"// {func_name} is undocumented. Please add documentation.\n")
        out.append(line)
        i += 1
        
    with open(filepath, 'w') as f:
        f.writelines(out)

for root, _, files in os.walk('go'):
    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go'):
            add_go_docs(os.path.join(root, file))
