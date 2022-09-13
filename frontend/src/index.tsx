import React, {useEffect, useState} from 'react'
import { createRoot } from 'react-dom/client'

import {Accordion, Stack, Tab, Tabs} from "react-bootstrap"
import 'bootstrap/dist/css/bootstrap.min.css'

import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import ChartDataLabels from 'chartjs-plugin-datalabels'

import { getFileSystems, getName } from './API'
import { FileSystem } from './FileSystemComponent'
import { LayeredFileSystemsComponent, LayerDatum } from './LayeredFileSystemsComponent'
import "./index.css"

ChartJS.register(ArcElement, Tooltip, ChartDataLabels, Legend)
ChartJS.overrides['pie'].plugins.legend.display = false

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
                        command: fileSystem.Command,
                        size: fileSystem.Size
                    }
                    layerData.push(layerDatum)
                }
            }
            layerData.sort(function(a, b){
                return b.rootDirId - a.rootDirId
            });
            setLayerData(layerData)
        }
        fetchName().catch(console.error)
        fetchFileSystems().catch(console.error)
    }, [])

    return (
        <Stack>
            <div className="bg-light border">
                <h1 className="center">{imageName}</h1>
            </div>
            <Tabs>
                <Tab eventKey="Image" title="Image">
                    <FileSystem rootDirId={imageRootId}></FileSystem>
                </Tab>
                <Tab eventKey="Layers" title="Layers">
                    <LayeredFileSystemsComponent layerData={layerData}></LayeredFileSystemsComponent>
                </Tab>
            </Tabs>
        </Stack>
    )
}

const container = document.getElementById('app')
const root = createRoot(container!)
root.render(<App/>)
