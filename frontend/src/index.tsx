import React, {MouseEvent, useEffect, useState, useRef} from 'react'
import { createRoot } from 'react-dom/client'

import Breadcrumb from 'react-bootstrap/Breadcrumb'
import {BreadcrumbItem} from "react-bootstrap"
import 'bootstrap/dist/css/bootstrap.min.css'

import FileTree from 'react-file-treeview'

import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Chart, getElementAtEvent } from 'react-chartjs-2'
import ChartDataLabels from 'chartjs-plugin-datalabels'

import { getDirData, getFileSystems, getName } from './API'

ChartJS.register(ArcElement, Tooltip, ChartDataLabels, Legend)
ChartJS.overrides['pie'].plugins.legend.position = 'bottom'

type LayerDatum = {
    rootDirId: number
    command: string
}

interface FileSystemProps {
    rootDirId: number
}

type DirData = {
    Id: number
    IsDir: boolean
    Name: string
    Size: number
    Files: DirData[]
}


const FileSystem: React.FC<FileSystemProps> = (props: FileSystemProps) => {

    const [dirStack, setDirStack] = useState<DirData[]>([])
    const pieChartRef = useRef<ChartJS>()

    const FILE_COLOR = "rgba(141, 210, 248, 0.8)"
    const DIRE_COLOR = "rgba(39, 17, 190, 0.28)"

    const copyDirStack = (): DirData[] => {
        return [...dirStack];
    }

    const getDirDataAndPushToDirStack = async (dirId: number) => {
        let dirData: DirData = await getDirData(dirId)
        let newDirStack = copyDirStack()
        if (dirData.Name === "/") {
            dirData.Name = " "
        }
        newDirStack.push(dirData)
        setDirStack(newDirStack)
    }

    const handleBreadcrumbClick = async () => {
        let newDirStack = copyDirStack()
        if (newDirStack.length < 2) {
            return
        }
        newDirStack.pop()
        setDirStack(newDirStack)
    }

    const handleSliceClick = async (event: MouseEvent<HTMLCanvasElement>) => {
        const { current: chart } = pieChartRef
        if (!chart) {
            return
        }
        const dir = dirStack.at(-1)
        if (dir === undefined) {
            return
        }
        const elem = getElementAtEvent(chart, event)[0]
        if (elem === undefined) {
            return
        }
        const clickedDataIndex: number = elem.index
        const fileClicked: DirData = dir.Files[clickedDataIndex]
        if (!fileClicked.IsDir) {
            return
        }
        await getDirDataAndPushToDirStack(fileClicked.Id).catch(console.error)
    }

    useEffect(() => {
        if (props.rootDirId < 0) {
            return
        }
        getDirDataAndPushToDirStack(props.rootDirId).catch(console.error)
    }, [props.rootDirId])

    const breadcrumbItems = []
    for (let dir of dirStack) {
        breadcrumbItems.push(
            <BreadcrumbItem key={dir.Id} active>
                {dir.Name}
            </BreadcrumbItem>
        )
    }

    const labels: string[] = []
    const percents: number[] = []
    const backgroundColors: string[] = []
    const curDir: DirData|undefined = dirStack.at(-1)
    if (curDir !== undefined) {
        for (let file of curDir.Files) {
            const label = file.Name
            const size = file.Size
            labels.push(label)
            percents.push(size)
            backgroundColors.push(file.IsDir ? DIRE_COLOR : FILE_COLOR)
        }
    }
    const pieChartData = {
        labels: labels,
        datasets: [
            {
                label: 'MB disk usage',
                data: percents,
                backgroundColor: backgroundColors,
                borderColor: "dark-blue",
                borderWidth: 1
            },
        ],
    }
    const options = {
        plugins: {
            datalabels: {
                formatter: (val: number) => {
                    if (val < 1) {
                        return ''
                    }
                    if (val > 1024*1024*1024) {
                        val = Math.round(val / 1024 / 1024 / 1024)
                        return val + ' GB'
                    }
                    if (val > 1024*1024) {
                        val = Math.round(val / 1024 / 1024)
                        return val + ' MB'
                    }
                    if (val > 1024) {
                        val = Math.round(val / 1024)
                        return val + ' KB'
                    }
                    return val + ' B'
                },
            }
        }
    }

    return (
        <div>
            <div onClick={async () => handleBreadcrumbClick()}>
                <Breadcrumb>
                    {breadcrumbItems}
                </Breadcrumb>
            </div>
            <div>

            </div>
            <div>
                <Chart ref={pieChartRef} type='pie' data={pieChartData} onClick={handleSliceClick} options={options}/>
            </div>
        </div>
    )
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

    return <div>
        <h1>{imageName}</h1>
        <FileSystem rootDirId={imageRootId}></FileSystem>
    </div>
}

const container = document.getElementById('app')
const root = createRoot(container!)
root.render(<App/>)
