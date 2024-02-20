const models = eval(require('fs').readFileSync('meta/models.json.js', 'utf8'));

const lib = {};
lib.setIndent = (indent, s) => {
    let minIndent = null;
    s.split('\n').forEach(line => {
        if (line.trim() !== '') {
            const leading = line.match(/^\s*/)[0];
            if (minIndent === null || leading.length < minIndent.length) {
                minIndent = leading;
            }
        }
    });

    let lines = s.split('\n');

    // remove leading and trailing blank lines
    while (lines.length > 0 && lines[0].trim() === '') {
        lines.shift();
    }
    while (lines.length > 0 && lines[lines.length - 1].trim() === '') {
        lines.pop();
    }

    return lines.map(line => {
        if (line.trim() === '') {
            return line;
        }
        return line.replace(minIndent, indent);
    }).join('\n');
}
lib.writefile = (filename, s) => {
    const HEADER = `// GENERATED(gen.js) - DO NOT EDIT BY HAND - See meta/models.json.js\n`;
    s = `${HEADER}\n${s}`;
    fs.writeFileSync(filename, s, 'utf8');
}

const outputters = {
    model_to_interface: m => {
        let s = `type ${m.name} interface {\n`
        const tab = '\t';
        for (let method in m.methods) {
            s += `${tab}${method}(${
                m.methods[method][0].map(x => x.join(' ')).join(', ')
            }) ${
                m.methods[method][1].length > 1
                    ? `(${m.methods[method][1].join(', ')})`
                    : m.methods[method][1][0]}\n`
        }
        s += `}\n`
        return s;
    }
};

const fs = require('fs');

for ( const model of models ) {
    {
        const filename = `${model.package}/${model.name}.go`;
        let s = `package ${model.package}\n\n`

        s += outputters.model_to_interface(model);

        lib.writefile(filename, s);
        console.log(`Wrote ${filename}`);
    }
    {
        const filename = `${model.package}/Base${model.name}.go`;
        let s = `package ${model.package}\n\n`

        if (model?.imports?.base) {
            s += `import (\n`
            for (let imp of model.imports.base) {
                s += `\t"${imp}"\n`
            }
            s += `)\n\n`
        }

        s += `type Base${model.name} struct {\n\t${model.name}\n}\n\n`

        for (let method in model.methods) {
            if ( model.methods[method].length >= 3 ) {
                s += `func (base *Base${model.name}) ${method}(${
                    model.methods[method][0].map(x => x.join(' ')).join(', ')
                }) ${
                    model.methods[method][1].length > 1
                        ? `(${model.methods[method][1].join(', ')})`
                        : model.methods[method][1][0]
                } {\n`
                s += lib.setIndent(`\t`, model.methods[method][2])
                s += `\n}\n`
            }
        }

        lib.writefile(filename, s);
        console.log(`Wrote ${filename}`);
    }
    {
        const filename = `${model.package}/Proxy${model.name}.go`;
        let s = `package ${model.package}\n\n`

        s += `type P_CreateProxy${model.name} struct {\n`
        s += `\tDelegate ${model.name}\n`
        s += `}\n\n`

        s += `type Proxy${model.name} struct {\n`
        s += `\tP_CreateProxy${model.name}\n`
        s += `}\n\n`

        s += `func CreateProxy${model.name}(params P_CreateProxy${model.name}) *Proxy${model.name} {\n`
        s += `\treturn &Proxy${model.name}{params}\n`
        s += `}\n\n`

        for (let method in model.methods) {
            s += `func (p *Proxy${model.name}) ${method}(${
                model.methods[method][0].map(x => x.join(' ')).join(', ')
            }) ${
                // model.methods[method][1][0]
                model.methods[method][1].length > 1
                    ? `(${model.methods[method][1].join(', ')})`
                    : model.methods[method][1][0]
            } {\n`
            s += `\treturn p.Delegate.${method}(${model.methods[method][0].map(x => x[0]).join(', ')})\n`
            s += `}\n`
        }

        lib.writefile(filename, s);
        console.log(`Wrote ${filename}`);
    }
}