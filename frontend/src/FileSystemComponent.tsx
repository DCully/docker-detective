import React, {MouseEvent, useEffect, useRef, useState} from "react"
import {Chart as ChartJS} from "chart.js"
import {getDirData} from "./API"
import {Chart, getElementAtEvent} from "react-chartjs-2"
import {BreadcrumbItem, Col, Container, ListGroupItem, Row} from "react-bootstrap"
import Breadcrumb from "react-bootstrap/Breadcrumb"
import ListGroup from "react-bootstrap/ListGroup"
import {rawBytesToReadableBytes} from './util'

export interface FileSystemProps {
    rootDirId: number
}

export type DirData = {
    Id: number
    IsDir: boolean
    Name: string
    Size: number
    Files: DirData[]
}

export const FileSystem: React.FC<FileSystemProps> = (props: FileSystemProps) => {

    const [dirStack, setDirStack] = useState<DirData[]>([])
    const pieChartRef = useRef<ChartJS>()

    const COLORS = [
        "rgba(141, 210, 248, 0.8)",
        "rgba(250, 148, 201, 0.8)",
        "rgba(39, 17, 190, 0.28)",
        "rgba(239, 244, 155, 0.8)",
        "rgba(119, 255, 132, 0.5)",
        "rgba(255, 141, 141, 0.8)"
    ]

    const copyDirStack = (): DirData[] => {
        return [...dirStack];
    }

    const getDirDataAndPushToDirStack = async (dirId: number) => {
        let dirData: DirData = await getDirData(dirId)
        let newDirStack = copyDirStack()
        newDirStack.push(dirData)
        if (newDirStack.length === 1) {
            // @ts-ignore
            newDirStack.at(0).Name = "/"
        } else {
            // @ts-ignore
            newDirStack.at(0).Name = " "
        }
        setDirStack(newDirStack)
    }

    const handleBreadcrumbClick = async () => {
        let newDirStack = copyDirStack()
        if (newDirStack.length < 2) {
            return
        }
        newDirStack.pop()
        if (newDirStack.length === 1) {
            // @ts-ignore
            newDirStack.at(0).Name = "/"
        }
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

    const handleListGroupItemClick = async (clickedDirData: DirData) => {
        await getDirDataAndPushToDirStack(clickedDirData.Id)
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

    const curDir: DirData|undefined = dirStack.at(-1)
    let files: DirData[] = []
    if (curDir !== undefined) {
        files = [...curDir.Files]
        files.sort(function(a, b){
            return b.Size - a.Size
        })
    }

    const labels: string[] = []
    const percents: number[] = []
    const backgroundColors: string[] = []
    for (let i = 0; i < files.length; i++) {
        const file = files[i]
        const label = file.Name
        const size = file.Size
        labels.push(label)
        percents.push(size)
        backgroundColors.push(COLORS[i % COLORS.length])
    }
    const pieChartData = {
        labels: labels,
        datasets: [
            {
                label: 'Disk usage',
                data: percents,
                backgroundColor: backgroundColors,
                borderColor: "gray",
                borderWidth: 1
            },
        ],
    }
    const options = {
        plugins: {
            datalabels: {
                formatter: rawBytesToReadableBytes,
            }
        }
    }

    let listGroupItems = []
    for (let i = 0; i < files.length; i++) {
        const file = files[i]
        if (file.IsDir && file.Size > 0) {
            listGroupItems.push(
                <ListGroupItem style={{backgroundColor: COLORS[i % COLORS.length]}}
                               key={file.Id}
                               action onClick={() => handleListGroupItemClick(file)}>
                    {rawBytesToReadableBytes(file.Size) + " - " + file.Name + "/"}
                </ListGroupItem>
            )
        } else {
            listGroupItems.push(
                <ListGroupItem style={{backgroundColor: COLORS[i % COLORS.length]}}
                               key={file.Id} disabled>
                    {rawBytesToReadableBytes(file.Size) + " - " + file.Name + (file.IsDir ? "/" : "")}
                </ListGroupItem>
            )
        }
    }

    return (
        <div>
            <Container>
                <div onClick={async () => handleBreadcrumbClick()}>
                    <Breadcrumb>
                        {breadcrumbItems}
                    </Breadcrumb>
                </div>
            </Container>
            <Container>
                <Row>
                    <Col>
                        <ListGroup style={{overflowY: 'scroll', maxHeight: window.innerHeight * 0.8}}>
                            {listGroupItems}
                        </ListGroup>
                    </Col>
                    <Col>
                        <Chart ref={pieChartRef} type='pie' data={pieChartData} onClick={handleSliceClick} options={options}/>
                    </Col>
                </Row>
            </Container>
        </div>
    )
}
