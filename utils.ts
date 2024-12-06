export function getTaskFile(task: string) {
    if (task.startsWith("file:")) {

        const readPermission = Deno.permissions.querySync({name: "read", path: task})
        if (readPermission.state === 'prompt') {
            const readRequest = Deno.permissions.requestSync({
                name: "read",
                path: task,
            })
            if (readRequest.state === 'granted') {
                return Deno.readTextFileSync(task)
            } else {
                console.error("Permission denied")
                Deno.exit(1)
            }
        } else if (readPermission.state === "granted") {
            return Deno.readTextFileSync(task)
        }
    } else if (task.startsWith("http")) {
        console.error("HTTP tasks are not supported yet")
        Deno.exit(1)
    } else {
        return task
    }
}

export function getExamplesFile(examples: string) {
    if (examples.startsWith("file:")) {
        const readPermission = Deno.permissions.querySync({name: "read", path: examples})
        if (readPermission.state === 'prompt') {
            const readRequest = Deno.permissions.requestSync({
                name: "read",
                path: examples,
            })
            if (readRequest.state === 'granted') {
                return Deno.readTextFileSync(examples)
            } else {
                console.error("Permission denied")
                Deno.exit(1)
            }
        } else if (readPermission.state === "granted") {
            return Deno.readTextFileSync(examples)
        }
    } else if (examples.startsWith("http")) {
        console.error("HTTP examples are not supported yet")
        Deno.exit(1)
    } else {
        return examples
    }
}

export function getFiles(files: string[]) {
    const filesContent = []
    for (const file of files) {
        const readPermission = Deno.permissions.querySync({name: "read", path: file})
        if (globalThis.isVariables) {
            console.log({file, readPermission})
        }
        if (readPermission.state === 'prompt') {
            const readRequest = Deno.permissions.requestSync({
                name: "read",
                path: file,
            })
            if (readRequest.state === 'granted') {
                const stat = Deno.statSync(file)
                const fileContent = Deno.readTextFileSync(file)
                if (globalThis.isVariables) {
                    console.log({fileContent, stat})
                }
                const fileName = file
                filesContent.push(`<${fileName}>\n${fileContent}`)
            } else {
                console.error("Permission denied")
                Deno.exit(1)
            }
        } else if (readPermission.state === 'granted') {
            const stat = Deno.statSync(file)
            const fileContent = Deno.readTextFileSync(file)
            if (globalThis.isVariables) {
                console.log({fileContent, stat})
            }
            const fileName = file
            filesContent.push(`<${fileName}>\n${fileContent}`)
        }
    }
    return filesContent.join("\n")
}

export function getApiGatewaySchemaFile(schema: string) {

    const readPermission = Deno.permissions.querySync({name: "read", path: schema})
    if (readPermission.state === 'prompt') {
        const readRequest = Deno.permissions.requestSync({
            name: "read",
            path: schema,
        })
        if (readRequest.state === 'granted') {
            return Deno.readTextFileSync(schema)
        } else {
            console.error("Permission denied")
            Deno.exit(1)
        }
    } else if (readPermission.state === "granted") {
        return Deno.readTextFileSync(schema)
    } else {
        return schema
    }

}
