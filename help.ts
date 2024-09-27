import { getConfigHelp } from "./config.ts"
import { mainHelp } from "./main.ts"

export function runHelp(subcommand?: string) {
    if (!subcommand) {
        console.log(mainHelp())
    } else if (subcommand === "config") {
        console.log(getConfigHelp())
    } else {
        console.log(`Unknown subcommand: ${subcommand}`)
        console.log(mainHelp())
    }
}
