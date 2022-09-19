import React from 'react'
import {Chart} from "react-chartjs-2";

export interface EfficiencyComponentProps {
    score: number
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
        labels: ["Useful", "Bloat"],
        datasets: [
            {
                label: 'Efficiency',
                data: [
                    props.score,
                    100 - props.score
                ],
                backgroundColor: [
                    "green",
                    "red"
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
            <h1 className="center"><b>{(100 - props.score) + "%"}</b> of this image's bytes are bloat</h1>
            <div>
                <Chart type='pie' data={data} options={options}/>
            </div>
            <div className="efficiency-p-div">
                <p>
                    An image's overall <b>efficiency score</b> derives from the percentage
                    of bytes in the layers of the Docker image which are ultimately
                    included in the resulting Docker image, after the image's layers have been
                    layered on top of each other.
                <br/><br/>
                    Generally, <b>lower bloat is better</b>. Try
                    to minimize the number of bytes in lower layers which are overwritten
                    or deleted in higher layers, to maximize efficiency when transporting your
                    Docker images across networks and storing them on disk.
                </p>
            </div>
        </div>
    )
}
