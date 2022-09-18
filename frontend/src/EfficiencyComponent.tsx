import React from 'react'
import {Chart} from "react-chartjs-2";

export interface EfficiencyComponentProps {
    score: number
}

const getEfficiencyPieSliceColor = (efficiencyScore: number): string => {
    if (efficiencyScore > 90) {
        return "green"
    } else if (efficiencyScore > 80) {
        return "yellow"
    }
    return "red"
}

export const getEfficiencyBadgeColor = (efficiencyScore: number): string => {
    if (efficiencyScore > 90) {
        return "success"
    } else if (efficiencyScore > 80) {
        return "warning"
    }
    return "danger"
}

export const EfficiencyComponent: React.FC<EfficiencyComponentProps> = (props: EfficiencyComponentProps) => {

    const data = {
        labels: ["Useful", "Waste"],
        datasets: [
            {
                label: 'Efficiency',
                data: [
                    props.score,
                    100 - props.score
                ],
                backgroundColor: [
                    getEfficiencyPieSliceColor(props.score),
                    getEfficiencyPieSliceColor(100 - props.score)
                ],
                borderColor: "black",
                borderWidth: 1
            },
        ],
    }
    const options = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            datalabels: {
                formatter: () => {return ''},
            }
        }
    }

    return (
        <div className="center">
            <h1 className="center">This image is <b>{props.score + "%"}</b> size efficient</h1>
            <div>
                <Chart type='pie' data={data} options={options}/>
            </div>
            <div className="efficiency-p-div">
                <p>
                    An image's overall <b>size efficiency score</b> derives from the percentage
                    of bytes in files in the layers of the Docker image which are ultimately
                    included in the resulting image file system, after Docker combines the
                    individual layers to create containers from the image.
                <br/><br/>
                    Generally, a <b>higher score is better</b> - try
                    to minimize the number of file bytes in lower layers which are overwritten
                    or deleted in higher layers, to maximize efficiency when transporting our
                    Docker images across networks and storing them on disk.
                </p>
            </div>
        </div>
    )
}
