import React, {useEffect, useState} from 'react'
import { createRoot } from 'react-dom/client'

import {Stack} from "react-bootstrap";
import 'bootstrap/dist/css/bootstrap.min.css'

import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import ChartDataLabels from 'chartjs-plugin-datalabels'

import { getDirData, getFileSystems, getName } from './API'
import { FileSystem } from './FileSystemComponent'
import "./index.css"

ChartJS.register(ArcElement, Tooltip, ChartDataLabels, Legend)
ChartJS.overrides['pie'].plugins.legend.display = false

type LayerDatum = {
    rootDirId: number
    command: string
}

const App: React.FC = () => {

    const[imageName, setImageName] = useState<string>('')
    const[imageRootId, setImageRootId] = useState<number>(-1)
    const[layerData, setLayerData] = useState<LayerDatum[]>([])

    useEffect(() => {
        const fetchName = async () => {
            const imageName = await getName()
            setImageName(imageName)
        }
        const fetchFileSystems = async () => {
            const fileSystems = await getFileSystems()
            const layerData: LayerDatum[] = []
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
            setLayerData(layerData)
        }
        fetchName().catch(console.error)
        fetchFileSystems().catch(console.error)
    }, [])

    return <Stack gap={3}>
        <div className="bg-light border">
            <h1 className="center">{imageName}</h1>
        </div>
        <FileSystem rootDirId={imageRootId}></FileSystem>
    </Stack>
}

const container = document.getElementById('app')
const root = createRoot(container!)
root.render(<App/>)
