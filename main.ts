#!/usr/bin/env -S deno run
import {parseArgs} from "jsr:@std/cli/parse-args"
import {runHelp} from "./help.ts"
import {handleConfig} from "./config.ts"
import {getExamplesFile, getFiles, getTaskFile} from "./utils.ts"
import {determineClient, listModels} from "./aiClients.ts"
import OpenAI from "openai/mod.ts"
import {generateCurl, generatePrompt} from "./prompt.ts"
import Anthropic from "npm:@anthropic-ai/sdk"
import {ChatCompletionCreateParamsBase} from "openai/resources/chat/completions.ts"
import {ChatCompletion} from "openai/resources/mod.ts"
import {MessageCreateParamsBase} from "npm:@anthropic-ai/sdk"



declare global {
  var isVerbose: boolean
}
if (import.meta.main) {
  const flags = parseArgs(Deno.args,
    {
      boolean: ["help", "verbose", "requiresLogin", "executeCommands"],
      alias: {
        help: "h",
        verbose: "v",
        model: "m",
        task: "t",
        files: "f",
        examples: "x",
        apiGatewaySchema: "a",
        apiKey: "k",
        endpoint: "e",
        requiresLogin: "r",
        executeCommands: "c"
      },
      collect: ["files"],
      string: ["model", "task", "files", "examples", "apiGatewaySchema", "apiKey", "endpoint",
      ],
      default: {
        verbose: false,
      }
    }
  )


  const subcommand = flags._[0]

  const isVerbose = flags.verbose || flags.v
  globalThis.isVerbose = isVerbose

  if (globalThis.isVerbose) {

    console.log(flags)
  }

  if (subcommand === "help" || flags.help) {
    const subSubCommand = flags._[1] as string
    runHelp(subSubCommand)
  } else if (subcommand === "config") {

    handleConfig(flags)

  } else if (subcommand === 'listModels') {
    const models = await listModels()
    const modelsOutput = {
      openai: {}, anthropic: {}
    }
    modelsOutput.openai = models.openai.data.map(model => model.id)
    modelsOutput.anthropic = models.anthropic.map(model => model.Name)
    console.log(modelsOutput)
  } else {
    const model = flags.model || flags.m
    if (!model) {
      console.error("Model name is required")
      Deno.exit(1)
    }
    const task = flags.task || flags.t || ""
    const files = flags.files || flags.f
    if (!files) {
      console.error("Files are required")
      Deno.exit(1)
    }
    const examples = flags.examples || flags.x || ""

    const apiGatewaySchema = flags.apiGatewaySchema || flags.a || ""
    const apiKey = flags.apiKey || flags.k || ""
    const endpoint = flags.endpoint || flags.e || ""
    const requiresLogin = flags.requiresLogin || flags.r || false
    const executeCommands = flags.executeCommands || flags.c || true

    const taskContent = getTaskFile(task)
    const filesContent = getFiles(files)
    const examplesContent = getExamplesFile(examples)
    if (globalThis.isVerbose) {
      console.log({taskContent, files, filesContent, examplesContent, executeCommands})
    }
    if (!taskContent || !filesContent) {
      console.error("Task and files are required")
      Deno.exit(1)
    }
    await generateCurls(model, taskContent, filesContent,
      apiGatewaySchema, apiKey, endpoint, examplesContent, requiresLogin, executeCommands)

  }
  Deno.exit(0)

}
export function mainHelp() {
  return `
  curlgen - A tool for generating curl commands

  Usage: curlgen <command> [options]

  Commands:
    config                    Manage configuration
    help                      Show help information

  Options:
    --help, -h                Show help
    --verbose, -v             Enable verbose mode
    --model, -m <model>       Specify the model to use (required)
    --task, -t <file>         Specify the task file
    --files, -f <files>       Specify the files to include (required)
    --examples, -e <file>     Specify the examples file
    --apiGatewaySchema, -a <schema>  Specify the API Gateway schema
    --apiKey, -k <key>        Specify the API key
    --endpoint, -e <url>      Specify the API endpoint
    --requiresLogin, -r       Specify if the API requires login

  For more information on a specific command, run:
    curlgen <command> --help
  `
}
async function generateCurls(model: string, taskContent: string,
  filescontent: string, apiGatewaySchema: string,
  apiKey: string, endpoint: string, examplesContent?: string, requiresLogin?: boolean, executeCommands?: boolean) {
  const {modelName, client} = await determineClient(model)


  const prompt = generatePrompt(taskContent, filescontent,
    examplesContent ?? "", apiGatewaySchema, requiresLogin ?? false)
  if (client instanceof OpenAI) {
    if (globalThis.isVerbose) {
      console.log({client: "openai"})
    }
    // Generate curls with OpenAI client
    await generateCurlsWithOpenAI(client, modelName, prompt, apiKey, endpoint, executeCommands ?? true)
    // console.log({curlCommands})
  } else if (client instanceof Anthropic) {
    // Generate curls with Anthropic client
    await generateCurlsWithAnthropic(client, modelName, prompt, apiKey, endpoint, executeCommands ?? true)

  } else {
    throw new Error("Unsupported client type")
  }


}

async function generateCurlsWithAnthropic(client: Anthropic, model: string, prompt: string, apiKey: string, endpoint: string, executeCommands: boolean) {

  const messageParams: MessageCreateParamsBase = {
    max_tokens: 1024,
    model,
    messages: [{
      role: "user",
      content: prompt
    }],
    tools: [
      {
        name: "generateCurlCommands",
        description: "Generate curl commands for testing an API endpoint",
        input_schema: {
          type: "object",
          properties: {
            commands: {
              type: "array",
              description: "Array of curl commands which covers all the test cases",
              items: {
                type: "object",
                properties: {
                  command: {
                    type: "string",
                    description: "Curl command to test an API endpoint"
                  },
                  explanation: {
                    type: "string",
                    description: "Explanation of the curl command"
                  }
                },
                required: ["command", "explanation"]
              }
            }
          },
          required: ["commands"]
        }
      }
    ]
  }
  if (globalThis.isVerbose) {
    console.log({messageParams: JSON.stringify(messageParams, null, 2)})
  }
  const curlCommands = await client.messages.create(messageParams)
  if (globalThis.isVerbose) {
    console.log({curlCommands: JSON.stringify(curlCommands, null, 2)})
  }
  const contentMessages = curlCommands.content
  for (const message of contentMessages) {
    if (message.type === "text") {
      console.log(message.text)
    } else if (message.type === "tool_use") {
      const commands = message.input ? message.input['commands'] as Array<{command: string, explanation: string}> : []

      const response = await runCurlsAndReturnResult(commands.map(command => command.command), endpoint, apiKey, executeCommands)
      if (globalThis.isVerbose) {

        console.log({response})
      }
      messageParams.messages.push({
        role: "assistant",
        content: curlCommands.content
      })
      messageParams.messages.push({
        role: "user",
        content: [{
          type: "tool_result", tool_use_id: message.id, content: `We were able to run the curls with following response: ${response.join("\n")}`

        }]
      })
      if (globalThis.isVerbose) {

        console.log({messageParams})
      }
    }
    const functionResponse = await client.messages.create(messageParams)
    if (functionResponse.stop_reason === "end_turn") {
      console.log(`The curl commands have been generated successfully`)
      console.log(functionResponse.content)
    }

  }

  return curlCommands
}

async function generateCurlsWithOpenAI(client: OpenAI, model: string, taskContent: string, apiKey: string,
  endpoint: string, executeCommands: boolean) {
  const chatParams: ChatCompletionCreateParamsBase = {
    model: model,
    messages: [
      {
        role: "user", content: taskContent
      }
    ],
    tools: [
      {
        type: "function",
        function: {
          name: "generateCurlCommands",
          description: "Generate curl commands for testing an API endpoint",
          parameters: {
            type: "object",
            properties: {
              commands: {
                type: "array",
                description: "Array of curl commands which covers all the test cases",
                items: {
                  type: "object",
                  properties: {
                    command: {
                      type: "string",
                      description: "Curl command to test an API endpoint"
                    },
                    explanation: {
                      type: "string",
                      description: "Explanation of the curl command"
                    }
                  },
                  required: ["command", "explanation"]
                }
              }
            },
            required: ["commands"]
          }
        }
      }
    ]
  }


  const curlCommands = await sendToOpenAi(model, chatParams, taskContent, client)
  const message = curlCommands.choices[0].message
  if (globalThis.isVerbose) {
    console.log({message})
  }
  if (curlCommands.choices[0].finish_reason === "tool_calls") {
    const toolCalls = curlCommands.choices[0].message.tool_calls
    if (toolCalls) {

      const extractedToolCalls = toolCalls.map(toolCall => ({
        id: toolCall.id,
        functionName: toolCall.function.name,
        arguments: JSON.parse(toolCall.function.arguments)
      }))
      for (const toolCall of extractedToolCalls) {
        if (toolCall.functionName === "generateCurlCommands") {

          const response = await runCurlsAndReturnResult(toolCall.arguments.commands.
            map(cmd => cmd.command), endpoint, apiKey, executeCommands)
          const id = toolCall.id
          if (globalThis.isVerbose) {

            console.log({response})
          }
          chatParams.messages.push(curlCommands.choices[0].message)
          chatParams.messages.push({
            role: "tool",
            tool_call_id: id,
            content: `We were able to run the curls with following response: ${response.join("\n")}`
          })



        }
        if (globalThis.isVerbose) {
          console.log({chatParams})
        }
        const functionResponse = await sendToOpenAi(model, chatParams, taskContent, client)
        if (functionResponse.choices[0].finish_reason === "stop") {
          console.log(`The curl commands have been generated successfully`)
          console.log(functionResponse.choices[0].message.content)
        }
        if (globalThis.isVerbose) {
          console.log({extractedToolCalls})
        }
      }
    }
  }
  return curlCommands
}

async function sendToOpenAi(model: string, chatParams: ChatCompletionCreateParamsBase, taskContent: string, client: OpenAI) {
  if (model.startsWith("o1")) {
    const userMessage = chatParams.messages.find(m => m.role === "user")
    const functionSchema = chatParams.tools ? JSON.stringify(chatParams.tools[0].function) : ""
    if (userMessage) {
      userMessage.content = `${taskContent}, your response should be an array of curl commands which should follow the schema: ${functionSchema}`
    }
    delete chatParams.tools
  }
  if (globalThis.isVerbose) {
    console.log({chatParams: JSON.stringify(chatParams, null, 2)})
  }
  const curlCommands = await client.chat.completions.create(chatParams) as ChatCompletion
  return curlCommands
}

async function runCurlsAndReturnResult(curlCommands: string[], endpoint: string, apiKey: string, executeCommands: boolean) {
  const results = []
  if (globalThis.isVerbose) {
    console.log({curlCommands, endpoint, apiKey, executeCommands})
  }
  for (const curlCommand of curlCommands) {
    let result = ``
    const commandWithoutFirstWord = generateCurl(curlCommand, endpoint, apiKey)
    if (globalThis.isVerbose) {
      console.log({commandWithoutFirstWord})
    }
    console.log(`Running: ${commandWithoutFirstWord}`)
    if (!executeCommands) {
      result = curlCommand
      results.push(result)
    } else {
      const command = new Deno.Command('sh', {
        args: ['-c', commandWithoutFirstWord]
      })
      const {code, stdout, stderr} = await command.output()
      const error = new TextDecoder().decode(stderr)
      const output = new TextDecoder().decode(stdout)

      if (code === 0) {

        result =
          `The curl command: ${curlCommand} returned the following output: ${output}`
        if (globalThis.isVerbose) {
          console.log({result})
        }
      } else {
        result = `The curl command: ${curlCommand} returned the following error: ${error}`
        console.error(`Error running curl command: ${curlCommand} and the following error: ${error}`)
      }
      results.push(result)
    }
  }
  return results
}


