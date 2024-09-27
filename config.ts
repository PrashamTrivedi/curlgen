


export const ANTHROPIC_KEY = "anthropic_api_key"
export const OPENAI_KEY = "openai_key"

export function readConfig(configKey?: string) {
    const configPath = createConfigPath()
    const config = JSON.parse(Deno.readTextFileSync(configPath))
    return configKey ? config[configKey] : config

}

function writeConfig(config: Record<string, string>) {
    const configPath = createConfigPath()
    Deno.writeTextFileSync(configPath, JSON.stringify(config, null, 2))
}

function resetConfig() {
    const configPath = createConfigPath()
    Deno.removeSync(configPath)
}

function createConfigPath() {
    const homeDir = Deno.env.get("HOME") || Deno.env.get("USERPROFILE")
    const configPath = `${homeDir}/.curlgen/config.json`
    return configPath
}

export function handleConfig(flags: Record<string, string | boolean | unknown>) {
    if (globalThis.isVerbose) {

        console.log(flags)
    }
    const subSubCommand = flags._[1]
    if (subSubCommand === "read") {
        const configKey = flags._[2] || flags.configKey
        const config = readConfig(configKey)
        console.log(config)
    } else if (subSubCommand === "write") {
        const configKey = flags._[2] || flags.configKey
        const configValue = flags._[3] || flags.configValue
        const config = readConfig()
        config[configKey] = configValue
        writeConfig(config)
    } else if (subSubCommand === "reset") {
        resetConfig()
    }
}

export function getConfigHelp(): string {
    return `
  Usage: curlgen config <subcommand> [configKey] [configValue]
  Subcommands:
    read [configKey]          Read configuration
    write [configKey] [configValue] Write configuration
    reset                     Reset configuration
  `
}