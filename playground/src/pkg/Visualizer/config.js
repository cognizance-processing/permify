function GraphOptions() {
    return  {
        autoResize: true,
        clickToUse: false,
        height: '90%',
        width: '100%',
        layout: {
            hierarchical: {
                enabled: false,
                direction: "UD",
                sortMethod: "hubsize",
                shakeTowards: "roots",
                levelSeparation: 150,
                nodeSpacing: 150,
                treeSpacing: 200
            }
        },
        interaction: {
            tooltipDelay: 10000,
            navigationButtons: true,
            keyboard: false,
            hover: true,
            multiselect: true,
            hoverConnectedEdges: false
        },
        physics: {
            forceAtlas2Based: {
                gravitationalConstant: -26,
                centralGravity: 0.005,
                springLength: 250,
                springConstant: 0.18,
                avoidOverlap: 1.5
            },
            maxVelocity: 30,
            solver: "forceAtlas2Based",
            timestep: 0.25,
            stabilization: {
                enabled: true,
                iterations: 1000,
                updateInterval: 25
            }
        },
        nodes: {
            fixed: {
                x: false,
                y: false
            },
            color: {
                hover: {
                    border: "#8246FF",
                    background: "#8246FF",
                }
            },
            font: {
                color: '#ffffff',
                size: 20
            },
            shape: "dot",
            size: 25,
            scaling: {
                min: 10,
                max: 60,
                label: {
                    enabled: true,
                    min: 20,
                    max: 32
                }
            }
        },
        edges: {
            hoverWidth: 1,
            arrows: {
                to: {
                    enabled: true,
                    scaleFactor: 1,
                    type: "arrow"
                }
            },
            color: {
                color: "#8246FF",
                highlight: "#8246FF",
                hover: "#8246FF",
                inherit: 'from'
            },
            font: {
                size: 20,
                color: "white",
                strokeWidth: 6,
                strokeColor: "#141517"
            },
            width: 3,
            smooth: true
        },
        groups: {
            entity: {
                color: { background: "#6318FF", border: "#6318FF" },
                scaling: { min: 20 },
                shape: "dot",
                size: 30,
            },
            relation: {
                color: { background: "#93F1EE", border: "#93F1EE" },
                scaling: { min: 10 },
                shape: "dot",
                size: 20,
            },
            permission: {
                borderRadius: "1px",
                color: { background: "#5bcc63", border: "#5bcc63" },
                shapeProperties: { borderDashes: false },
                shape: "box",
                size: 20,
            },
            logic: {
                color: { background: "#e53472", border: "#e53472" },
                shapeProperties: { borderDashes: false },
                shape: "icon",
                icon: {
                    face: "FontAwesome",
                    code: "\uf286",
                    size: 50,
                    color: "#e53472"
                },
                size: 15,
            },
        },
    };
}

export default GraphOptions
