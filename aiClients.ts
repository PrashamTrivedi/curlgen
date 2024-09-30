import OpenAI from "openai"
import Anthropic from "npm:@anthropic-ai/sdk"
import {ANTHROPIC_KEY, OPENAI_KEY, readConfig} from "./config.ts"
import {Model} from "openai/resources/models.ts"


let openAiClient: OpenAI | null = null
let anthropicClient: Anthropic | null = null

function initClients() {


    if (!openAiClient || !anthropicClient) {

        const openAiKey = readConfig(OPENAI_KEY)
        const anthropicKey = readConfig(ANTHROPIC_KEY)

        if (!openAiKey || !anthropicKey) {
            throw new Error("API keys are missing or invalid")
        }

        if (globalThis.isVariables) {
            console.log({openai: maskKey(openAiKey), anthropic: maskKey(anthropicKey)})
        }
        openAiClient = new OpenAI({
            apiKey: openAiKey
        })

        anthropicClient = new Anthropic({
            apiKey: anthropicKey
        })
    } else {

        if (globalThis.isVerbose) {
            console.log("Clients already initialized")
        }
    }
}

const anthropicModels = [{
    ApiName: "claude-3-5-sonnet-20240620",
    Name: "claude-3.5-sonnet",
},
{
    ApiName: "claude-3-opus-20240229",
    Name: "claude-3-opus",
},
{
    ApiName: "claude-3-sonnet-20240229",
    Name: "claude-3-sonnet",
},
{
    ApiName: "claude-3-haiku-20240307",
    Name: "claude-3-haiku",
},
]




const modelToClientMap: Record<string, OpenAI | Anthropic> = {}

export async function listModels() {
    initClients()
    const openAiModels = await initModelToClientMap()
    return {
        openai: openAiModels,
        anthropic: anthropicModels
    }
}

async function initModelToClientMap() {
    const openAiModels = await openAiClient.models.list()
    openAiModels.data.forEach((model: Model) => {
        modelToClientMap[model.id] = openAiClient
    })
    anthropicModels.forEach((model: {ApiName: string; Name: string}) => {
        modelToClientMap[model.Name] = anthropicClient
        modelToClientMap[model.ApiName] = anthropicClient
    })
    return openAiModels
}


export async function determineClient(modelName: string): {modelName: string, client: OpenAI | Anthropic} {
    initClients()
    if (!Object.keys(modelToClientMap).length) {
        await initModelToClientMap()
    }
    const client = modelToClientMap[modelName]
    if (client instanceof OpenAI) {
        return {modelName, client}
    } else {
        const anthropicModel = anthropicModels.find(model => model.Name === modelName)
        if (anthropicModel) {
            return {modelName: anthropicModel.ApiName, client}
        } else {
            throw new Error(`Model ${modelName} not found in Anthropic models`)
        }
    }

}

function maskKey(key: string): string {
    if (key.length <= 8) {
        return '*'.repeat(key.length)
    }
    return key.slice(0, 4) + '*'.repeat(key.length - 8) + key.slice(-4)
}
