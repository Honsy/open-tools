
export interface RouteItem {
    url:string;
    key:string;
    special?:boolean;
    component?:Function;
}
export interface TemplateData{
    title:string;
    data:string;
    ssrData:SSRData;
    state:string;
}

export interface SSRData{
    html:string
}

// interface ComponentProps {
//     url:string,
//     data:object
// }

// interface SSRProps {
//     componentName:string,
//     data:ComponentProps
// }

export default {}