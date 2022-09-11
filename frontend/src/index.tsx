import React, {useEffect, useState} from 'react';
import { createRoot } from 'react-dom/client';

// import Accordion from 'react-bootstrap/Accordion';
// import AccordionItem from "react-bootstrap/AccordionItem";
//
// import Tab from 'react-bootstrap/Tab';
// import Tabs from 'react-bootstrap/Tabs';
//

import Breadcrumb from 'react-bootstrap/Breadcrumb';

import 'bootstrap/dist/css/bootstrap.min.css';

import { getDirData, getFileSystems, getName } from "./API";

import "./index.css"
import {BreadcrumbItem} from "react-bootstrap";

type LayerDatum = {
    rootDirId: number;
    command: string;
}

interface FileSystemTreeMapProps {
    rootDirId: number;
}

type DirData = {
    Id: number
    IsDir: boolean
    Name: string
    Size: number
    Files: DirData[]
}

const FileSystemTreeMap: React.FC<FileSystemTreeMapProps> = (props: FileSystemTreeMapProps) => {

    // TODO - combine these two sources of state into one
    const [dirData, setDirData] = useState<DirData>({Id: -1, IsDir: false, Name: ' ', Size: 0, Files: []});
    const [dirStack, setDirStack] = useState<DirData[]>([dirData]);

    // this component will render the tree map
    // every time the user clicks a section in the tree map,
    // the onClick handler will fetch new dir data and feed
    // it into setDirData, which will trigger this component
    // getting re-rendered, drawing the new tree map.

    const updateDirDataFromApi = async (dirId: number) => {
        if (dirId < 0) {
            return
        }
        const rootDirData: DirData = await getDirData(dirId);
        setDirData(rootDirData);
    }

    const handleBreadcrumbClick = async () => {

        // copy the dir stack, removing the top element of the stack
        let newDirStack: DirData[] = [];
        for (const d of dirStack) {
            newDirStack.push(d);
        }
        newDirStack.pop();

        // fetch dir data for the new top of the stack
        const stackTop: DirData|undefined = newDirStack.at(-1);
        if (stackTop !== undefined) {
            await updateDirDataFromApi(stackTop.Id);
        }

        // re-render the breadcrumb
        setDirStack(newDirStack);
    }

    const handleRectClick = async (file: DirData) => {
        if (!file.IsDir) {
            console.log('clicked a file - ignoring rect click');
            return
        }
        await updateDirDataFromApi(file.Id);
        let newDirStack: DirData[] = [];
        for (const d of dirStack) {
            newDirStack.push(d);
        }
        newDirStack.push(file);
        setDirStack(newDirStack);
    }

    useEffect(() => {
        updateDirDataFromApi(props.rootDirId).catch(console.error);
    }, [props.rootDirId]);

    const rects = []
    const totalSize: number = dirData.Size;
    const files = dirData.Files;
    files.sort(function(a, b){
        return b.Size - a.Size
    });
    let consumedPercent: number = 0;
    for (let file of files) {
        const pctSize: number = ((file.Size / totalSize) * 100.0);
        rects.push(
            <g key={file.Id + "-g"}>
                <rect key={file.Id}
                    x={consumedPercent.toString() + "%"}
                    y="0%"
                    width={pctSize.toString() + "%"}
                    height="100%"
                    fill="grey" onClick={async () => handleRectClick(file)}>
                </rect>
                <text key={file.Id + "-text"}
                      x={(consumedPercent + 1).toString() + "%"}
                      y="5%"
                      fontSize="4">{file.Name + " - " + Math.round(file.Size / 1024 / 1024) + "MB"}
                </text>
            </g>
        )
        consumedPercent = consumedPercent + pctSize + 0.25;
    }

    const breadcrumbItems = []
    for (let dir of dirStack) {
        breadcrumbItems.push(
            <BreadcrumbItem key={dir.Id} active>
                {dir.Name}
            </BreadcrumbItem>
        )
    }

    return (
        <div>
            <div onClick={async () => handleBreadcrumbClick()}>
                <Breadcrumb>
                    {breadcrumbItems}
                </Breadcrumb>
            </div>
            <svg viewBox="0 0 200 80" preserveAspectRatio="xMidYMid meet">
                {rects}
            </svg>
        </div>
    )
}

const App: React.FC = () => {

    const[imageName, setImageName] = useState<string>('');
    const[imageRootId, setImageRootId] = useState<number>(-1);
    const[layerData, setLayerData] = useState<LayerDatum[]>([]);

    useEffect(() => {
        const fetchName = async () => {
            const imageName = await getName();
            setImageName(imageName)
        }
        const fetchFileSystems = async () => {
            const fileSystems = await getFileSystems();
            const layerData: LayerDatum[] = [];
            for (let fileSystem of fileSystems) {
                if (fileSystem.Name === "image") {
                    setImageRootId(fileSystem.RootDirectoryId)
                } else {
                    const layerDatum: LayerDatum = {
                        rootDirId: fileSystem.RootDirectoryId,
                        command: fileSystem.Command
                    }
                    layerData.push(layerDatum)
                }
            }
            layerData.sort(function(a, b){
                return a.rootDirId - b.rootDirId
            });
            setLayerData(layerData);
        }
        fetchName().catch(console.error);
        fetchFileSystems().catch(console.error);
    }, []);

    return <div>
        <h1>{imageName}</h1>
        <FileSystemTreeMap rootDirId={imageRootId}></FileSystemTreeMap>
    </div>
}

const container = document.getElementById('app')
const root = createRoot(container!)
root.render(<App/>)
