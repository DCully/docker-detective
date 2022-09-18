import React from "react"
import {Accordion, Badge, Col, Container, Row} from "react-bootstrap"
import {FileSystem} from "./FileSystemComponent"
import {rawBytesToReadableBytes} from './util'

export type LayerDatum = {
    rootDirId: number
    command: string
    size: number
}

interface LayeredFileSystemsProps {
    layerData: LayerDatum[]
}

export const LayeredFileSystemsComponent: React.FC<LayeredFileSystemsProps> = (props: LayeredFileSystemsProps) => {

    const accordionItems = []
    for (const layer of props.layerData) {
        accordionItems.push(
            <Accordion.Item eventKey={layer.rootDirId.toString()}>
                <Accordion.Header>
                    <Container fluid>
                        <Row>
                            <Col className="col-1">
                                <Badge bg="secondary">{rawBytesToReadableBytes(layer.size)}</Badge>
                            </Col>
                            <Col>
                                {layer.command}
                            </Col>
                        </Row>
                    </Container>
                </Accordion.Header>
                <Accordion.Body>
                    <FileSystem rootDirId={layer.rootDirId}></FileSystem>
                </Accordion.Body>
            </Accordion.Item>
        )
    }

    return <Accordion flush>{accordionItems}</Accordion>
}
