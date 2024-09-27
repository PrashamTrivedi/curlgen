import Template from "https://deno.land/x/template@v0.1.0/mod.ts"
export const basePrompt =
    `You are an expert API tester. Your task is to generate curl commands for testing an API endpoint based on the following information:

Generate appropriate curl commands for testing the API endpoint based on the following information:

Task Definition:{{taskContent}}
Updated Code: {{filesContent}}
Api Gateway Schema: {{apiGatewaySchema}}
Examples: {{examplesContent}}
Additional Information: {{additionalInfo}}
Requires Login: {{requiresLogin}}

Consider the following when generating curl commands:

1. Use the API Gateway Request Model to structure the request body (if provided).
2. Incorporate any authentication or headers required by the API.
3. Include examples for different HTTP methods (GET, POST, PUT, DELETE) if applicable.
4. Provide variations of the curl commands to test different scenarios or edge cases.
5. Include any necessary query parameters or path variables.
6. Use the placeholder {{API_URL}} for the API URL.
7. Use the placeholder {{API_KEY}} for the API Key or Authorization Header if the curl requires login.

Generate and run a set of curl commands that thoroughly test the API endpoint described above, using the placeholders for API URL and API Key.`


export function generatePrompt(taskContent: string, filesContent: string, examplesContent: string,
    apiGatewaySchema: string,
    requiresLogin: boolean,
    additionalInfo?: string) {
    const template = new Template()
    return template.render(basePrompt, {
        taskContent,
        filesContent,
        examplesContent,
        apiGatewaySchema,
        additionalInfo,
        requiresLogin
    })
}


export function generateCurl(command: string, endpoint: string, apiKey: string): string {
    const curlTemplate = new Template()
    return curlTemplate.render(command, {
        API_URL: endpoint,
        API_KEY: apiKey
    })
}