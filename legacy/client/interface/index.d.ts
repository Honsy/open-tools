export {}

declare global {
    interface Window {
        STATE_FROM_SERVER: any
    }
}

export interface ToolObj{
    title:string
    data:ToolItemObj[]
}

export interface ToolItemObj{
    title:string
    href:string
    color:string
}