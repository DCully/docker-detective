import React, {useEffect, useState} from 'react'
import { createRoot } from 'react-dom/client'

import {Badge, Card, Stack, Tab, Tabs} from "react-bootstrap"
import 'bootstrap/dist/css/bootstrap.min.css'

import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import ChartDataLabels from 'chartjs-plugin-datalabels'

import { getFileSystems, getName } from './API'
import { FileSystem } from './FileSystemComponent'
import { LayeredFileSystemsComponent, LayerDatum } from './LayeredFileSystemsComponent'
import "./index.css"
import {rawBytesToReadableBytes} from "./util";
import {EfficiencyComponent, getEfficiencyBadgeColor} from "./EfficiencyComponent";

ChartJS.register(ArcElement, Tooltip, ChartDataLabels, Legend)
ChartJS.overrides['pie'].plugins.legend.display = false

const App: React.FC = () => {

    const[imageName, setImageName] = useState<string>('')
    const[imageRootId, setImageRootId] = useState<number>(-1)
    const[totalImageSizeBytes, setTotalImageSizeBytes] = useState<number>(0)
    const[layerData, setLayerData] = useState<LayerDatum[]>([])

    useEffect(() => {
        const fetchName = async () => {
            const imageName = await getName()
            setImageName(imageName)
        }
        const fetchFileSystems = async () => {
            const fileSystems = await getFileSystems()
            const layerData: LayerDatum[] = []
            let totalLayerSize: number = 0
            for (let fileSystem of fileSystems) {
                if (fileSystem.Name === "image") {
                    setImageRootId(fileSystem.RootDirectoryId)
                    setTotalImageSizeBytes(fileSystem.Size)
                } else {
                    const layerDatum: LayerDatum = {
                        rootDirId: fileSystem.RootDirectoryId,
                        command: fileSystem.Command,
                        size: fileSystem.Size
                    }
                    layerData.push(layerDatum)
                    totalLayerSize = totalLayerSize + fileSystem.Size
                }
            }
            setLayerData(layerData)
        }
        fetchName().catch(console.error)
        fetchFileSystems().catch(console.error)
    }, [])

    const getLayersTotalSize = (): number => {
        let sum: number = 0
        for (const layer of layerData) {
            sum = sum + layer.size
        }
        return sum
    }

    const efficiencyScore: number = Math.round(getLayersTotalSize() / totalImageSizeBytes * 100)
    const efficiencyScoreStr: string =  efficiencyScore.toString() + "%"

    return (
        <Stack>
            <div className="bg-light border">
                <h1 className="center">{imageName}</h1>
            </div>
            <Tabs defaultActiveKey="Efficiency">
                <Tab eventKey="Image" title={
                    <React.Fragment>
                        <Badge bg="secondary">{rawBytesToReadableBytes(totalImageSizeBytes)}</Badge>
                        <div>Image File System</div>
                    </React.Fragment>
                }>
                    <FileSystem rootDirId={imageRootId}></FileSystem>
                </Tab>
                <Tab eventKey="Efficiency" title={
                    <React.Fragment>
                        <Badge bg={getEfficiencyBadgeColor(efficiencyScore)}>{efficiencyScoreStr}</Badge>
                        <div>Space Efficiency</div>
                    </React.Fragment>
                }>
                    <Card>
                        <Card.Body className="center">
                            <EfficiencyComponent score={efficiencyScore}></EfficiencyComponent>
                        </Card.Body>
                    </Card>
                </Tab>
                <Tab eventKey="Layers" title={
                    <React.Fragment>
                        <Badge bg="secondary">{rawBytesToReadableBytes(getLayersTotalSize())}</Badge>
                        <div>Layer File Systems</div>
                    </React.Fragment>
                }>
                    <LayeredFileSystemsComponent layerData={layerData}></LayeredFileSystemsComponent>
                </Tab>
            </Tabs>
        </Stack>
    )
}

const container = document.getElementById('app')
const root = createRoot(container!)
root.render(<App/>)
