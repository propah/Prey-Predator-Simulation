
const SCALING = 5;

const names = ["Jean", "Marie", "Pierre", "Huguette", "Roger", "Gertrude", "Gérard", "Ginette", "René", 
                "Claudine", "Germaine", "Solange", "Marcel", "Lucienne", "Albert", "Josette", "Bernard",
                "Paulette", "Raymond", "Léon", "Louis", "Alberte", "José", "Marion", "Marcelle", "Jules",
                "Gertrudette", "Gilberte", "Fernand", "Lucette", "Armand", "Bernadette", "Roland",
                "Ronald", "Marguerite", "Lea", "Arnaud", "Lucerine", "Norbert", "Monique", "Edouard",
                "Eude", "Claudette", "Antoine", "André", "Paul", "Gilbert", "Renette", "Astier", "Simone", 
                "Hervé", "Hugues", "Gaston", "Emile", "Elizabeth", "Georges", "Odette", "Claude", 
                "Berthe", "Lucien", "Lucérine", "Thérèse", "Léontine", "Arlette", "Bernadette", "Fernandine", 
                "Arnaudine", "Félix", "Joséphine", "Henriette", "Henri", "Philibert", "Gilles", "Simon",
                "Mélissandre", "Alexis", "Alexandre", "Georgette", "Philippe", "Fantine", "Bernard"]
const names_length = names.length;

function getRandomName() {
    let name = ""

    name += names[Math.floor(Math.random() * names_length)];
    name += "-";
    name += names[Math.floor(Math.random() * names_length)];

    return name
}

function scaleAgent(newAgent) {
    newAgent.pos[0] *= SCALING;
    newAgent.pos[1] *= SCALING;
    if (newAgent.vel) {
        newAgent.vel[0] *= SCALING;
        newAgent.vel[1] *= SCALING;
    }
    if (newAgent.raysValues) {
        newAgent.raysValues = newAgent.raysValues.map(rayValue => {
            let scaledValue = rayValue * SCALING;
            if (rayValue < 0) {
                scaledValue = -scaledValue + 1000;
            }
            return scaledValue;
        });
    }
}

class Application {
    constructor(agentCount, height, width, cellSize, agentRadius, predatorRayAngleDeg, preyRayAngleDeg, predatorRayMaxLength, preyRayMaxLength) {
        this.agents = new Map();
        this.serverMessageCount = 0;
        this.socket = new WebSocket("ws://localhost:8080/ws");
        this.cellSize = cellSize;
        this.agentCount = agentCount;
        this.agentRadius = agentRadius;
        this.zoomScaleFactor = 1;
        this.zoomScaleMultiplier = 1.1;
        this.dragging = false;
        this.height = height;
        this.width = width;
        this.selectedAgentId = null;
        this.mousePos = {x: 0, y: 0};
        this.rayLines = null;
        this.directionLine = null;
        this.predatorRayAngleDeg = predatorRayAngleDeg
        this.preyRayAngleDeg = preyRayAngleDeg
        this.PredatorRayMaxLength = predatorRayMaxLength
        this.PreyRayMaxLength = preyRayMaxLength
    }

    initialize() {
        this.app = this.setUpPixiApp(window.innerWidth, window.innerHeight);
        this.agentTexture = this.generateAgentTexture(this.agentRadius);
        this.setUpContainers();
        this.grid = this.setUpGrid();
        this.socket.addEventListener("message", (event) => this.handleWebSocketMessage(event));
        this.setUpPeriodicUpdates();
        this.addMouseInteractions();
        this.addTickerUpdates();
        this.background = this.setUpBackgroundImage();
        this.sapinsImage = this.setUpSapinsImage();
    }

    setUpPixiApp(width, height) {
        const app = new PIXI.Application({ width, height, backgroundColor: 0x8BB164, antialias: true, });

        document.getElementById('canvas-container').appendChild(app.view);
        return app;
    }

    setUpBackgroundImage() {
        const background = new PIXI.Sprite.from('./img/map.png');

        background.anchor.set(0.085);
        background.width = this.width+1000;
        background.height = this.height+1000;

        this.particleContainer.addChild(background);
        return background;
    }

    setUpSapinsImage() {
        const sapinsImage = new PIXI.Sprite.from('./img/sapins.png');

        sapinsImage.anchor.set(0.085);
        sapinsImage.width = this.width+1000;
        sapinsImage.height = this.height+1000;

        this.particleContainer.addChild(sapinsImage);
        return sapinsImage;
    }

    generateAgentTexture(agentSize) {
        let graphics = new PIXI.Graphics();
        graphics.beginFill(0xFFFFFF);
        graphics.drawCircle(0, 0, agentSize);
        graphics.endFill();
        return this.app.renderer.generateTexture(graphics);
    }

    setUpContainers() {
        let particleContainer = new PIXI.Container();
        let neuralNetworkContainer = new PIXI.Container();

        this.particleContainer = particleContainer;
        this.neuralNetworkContainer = neuralNetworkContainer;
        this.neuralNetworkContainer.x = window.innerWidth * 0.9;
        this.neuralNetworkContainer.y = window.innerHeight / 2;

        this.app.stage.addChild(particleContainer);
        this.app.stage.addChild(neuralNetworkContainer);
    }

    setUpGrid() {
        let grid = new PIXI.Graphics();
        grid.lineStyle(SCALING / 2, 0x000000, 0.2); 
        for (let i = 0; i <= this.width; i += this.cellSize) {
            grid.moveTo(i, 0);
            grid.lineTo(i, this.height);
        }
        for (let i = 0; i <= this.height; i += this.cellSize) {
            grid.moveTo(0, i);
            grid.lineTo(this.width, i);
        }
        this.app.stage.addChild(grid);
        return grid;
    }

    handleWebSocketMessage(event) {
        this.serverMessageCount++;
        const data = JSON.parse(event.data);
        this.updateAgents(data.agents);
        this.updateInfos(data);
    }

    updateAgents(newAgents) {
        this.updateDebugInfo();
        const receivedAgentIds = new Set(newAgents.map(agent => agent.id));

        this.agents.forEach((agent, id) => {
            if (!receivedAgentIds.has(id)) {
                this.particleContainer.removeChild(agent);
                this.agents.delete(id);
                if (this.selectedAgentId === id) {
                    this.selectedAgentId = null;
                }
            }
        });

        newAgents.forEach((newAgent) => {
            scaleAgent(newAgent);
            let agent;
            let agid = this.agents.get(newAgent.id);

            if (newAgent.color === 'Red') {
                if (agid) agent = agid;
                else {
                    agent = new PIXI.Sprite.from('./img/phidippus.png');
                    agent.width = 50;
                    agent.height = 50;
                    agent.name = getRandomName();
                }
            } else {
                if (agid) agent = agid;
                else {
                    agent = new PIXI.Sprite.from('./img/drosophile.png');
                    agent.width = 40;
                    agent.height = 40;
                }
            }

            if (!this.agents.has(newAgent.id)) {
                this.particleContainer.addChild(agent);
                this.agents.set(newAgent.id, agent);
            }

            agent.anchor.set(0.5);
            agent.position.set(newAgent.pos[0], newAgent.pos[1]);
            if (this.selectedAgentId == newAgent.id) {
                agent.alpha = 1;
            } else {
                agent.alpha = 1;
            }

            if (newAgent.vel && newAgent.raysValues) {
                this.drawAgentRays(newAgent);
                this.drawDirectionLine(newAgent);
                this.drawNeuralNetwork(newAgent);
                this.fillJauges(newAgent);
            }

            if (newAgent.vel) {
                this.rotateAgent(newAgent, agent);
            }
        });

        document.getElementById("agentcount").innerHTML = this.agents.size.toString();
        this.particleContainer.removeChild(this.sapinsImage);
        this.particleContainer.addChild(this.sapinsImage);
    }

    rotateAgent(newAgent, agent) {
        let vel = newAgent.vel;
        const angle = Math.atan2(vel[1], vel[0]);
        agent.rotation = angle;
    }

    setUpPeriodicUpdates() {
        setInterval(() => {
            document.getElementById("framerate").innerHTML = `${this.app.ticker.FPS.toFixed(2)} FPS`;
        }, 100);

        setInterval(() => {
            document.getElementById("tickrate").innerHTML = `${this.serverMessageCount} ticks/s`;
            this.serverMessageCount = 0;
        }, 1000);
    }

    addMouseInteractions() {
        this.app.view.addEventListener('wheel', (e) => {
            e.preventDefault();
        
            const oldZoomScaleFactor = this.zoomScaleFactor;
            this.zoomScaleFactor = e.deltaY < 0 ? 
                this.zoomScaleFactor * this.zoomScaleMultiplier : 
                this.zoomScaleFactor / this.zoomScaleMultiplier;
            
            const centerX = this.app.renderer.width / 2;
            const centerY = this.app.renderer.height / 2;

            // on calcule la nouvelle position de l'écran après le zoom
            const newPosX = (this.particleContainer.x - centerX) * (this.zoomScaleFactor / oldZoomScaleFactor) + centerX;
            const newPosY = (this.particleContainer.y - centerY) * (this.zoomScaleFactor / oldZoomScaleFactor) + centerY;
        
            this.particleContainer.scale.set(this.zoomScaleFactor);
            this.particleContainer.position.set(newPosX, newPosY);

            this.grid.scale.set(this.zoomScaleFactor);
            this.grid.position.set(newPosX, newPosY);
        });

        this.app.view.addEventListener('mousedown', (e) => { 
            this.mousePos.x = e.clientX;
            this.mousePos.y = e.clientY;
            if (e.button === 0) {
                const mousePos = this.mousePos;
                mousePos.x -= this.particleContainer.x;
                mousePos.y -= this.particleContainer.y;
                let closestDistance = Infinity;
                let closestAgentId = null;
    
                this.agents.forEach((agent, id) => {
                    const dx = agent.x * this.zoomScaleFactor - mousePos.x;
                    const dy = agent.y * this.zoomScaleFactor - mousePos.y;
                    const distanceSquared = dx * dx + dy * dy; // Should use sqrt but we don't care about the real distance, we want only the closest
    
                    if (distanceSquared < closestDistance) {
                        closestDistance = distanceSquared;
                        closestAgentId = id;
                    }
                });
    
                if (closestAgentId !== null) {
                    fetch(`http://localhost:8080/selectAgent`, 
                        { "method": "POST", "body": '{ "agentID": ' + closestAgentId.toString() + "}" }, 
                        { "Content-Type": "application/json" }
                    )
                    this.highlightAgent(closestAgentId);
                    this.updateAgentInfo(closestAgentId);
                }
            }
            this.dragging = true;
        });
        this.app.view.addEventListener('mouseup', () => { this.dragging = false; });
        this.app.view.addEventListener('mousemove', (e) => {
            this.mousePos.x = e.clientX;
            this.mousePos.y = e.clientY;
            if (this.dragging) {
                this.particleContainer.x += e.movementX;
                this.particleContainer.y += e.movementY;
                this.grid.x += e.movementX;
                this.grid.y += e.movementY;
            }
        });
    }

    highlightAgent(agentId) {
        /*if (this.selectedAgentId !== null) {
            this.resetAgentAppearance(this.selectedAgentId);
        }*/
        const agent = this.agents.get(agentId);
        if (agent) {
            agent.alpha = 1; 
        } else {
            console.log("Agent not found");
        }
        this.selectedAgentId = agentId;
        this.selectedAgent = agent;
    }

    fillJauges(agent) {
        document.getElementById("nogen").innerHTML = agent.generation;
        document.getElementById("jauge_lifepoints").setAttribute("style", ("width:" + agent.lifepoints + "%"));
        document.getElementById("jauge_energy").setAttribute("style", ("width:" + agent.energy + "%"));
        document.getElementById("jauge_reproduction").setAttribute("style", ("width:" + agent.reproduction + "%"));
        document.getElementById("jauge_digestion").setAttribute("style", ("width:" + agent.digestion + "%"));
    }
    
    resetAgentAppearance(agentId) {
        const agent = this.agents.get(agentId);
        agent.alpha = 0.3;
    }

    updateAgentInfo(agentId) {
        if (agentId === null) {
            return;
        }
        const agent = this.agents.get(agentId);
        document.getElementById("agentid").innerHTML = agentId.toString();
        if (this.selectedAgent.name) {
            document.getElementById("agentname").innerHTML = "| " + this.selectedAgent.name;
        } else {
            document.getElementById("agentname").innerHTML = "";
        }
        document.getElementById("agentx").innerHTML = (agent.x/SCALING).toFixed(2);
        document.getElementById("agenty").innerHTML = (agent.y/SCALING).toFixed(2);
    }

    addTickerUpdates() {
        this.ticker = new PIXI.Ticker();
        this.ticker.add(() => this.updateDebugInfo());
        this.ticker.add(() => this.updateAgentInfo(this.selectedAgentId));
        this.ticker.start();
        console.log(this.ticker);
    }

    updateDebugInfo() {
        document.getElementById("mousepos").innerHTML = `X=${this.mousePos.x.toFixed(2)} Y=${this.mousePos.y.toFixed(2)}`;
        document.getElementById("zoomscale").innerHTML = this.zoomScaleFactor.toFixed(2);
        document.getElementById("gridpos").innerHTML = `X=${this.grid.x.toFixed(2)} Y=${this.grid.y.toFixed(2)}`;
    }

    updateInfos(data) {
        document.getElementById("elapsedtime").innerHTML = this.formatTime(data.elapsedtime);
        document.getElementById("tickscount").innerHTML = data.tickcounter;
        document.getElementById("phidicount").innerHTML = data.predatorcount;
        document.getElementById("preycount").innerHTML = data.preycount;
    }

    formatTime(elapsedtime) {
        let hh = Math.floor(elapsedtime / (1000*60*60));
        let mm = Math.floor((elapsedtime % (1000*60*60)) / (1000*60));
        let ss = Math.floor((elapsedtime % (1000 * 60)) / 1000);
        return `${hh}`.padStart(2,'0') + ':' 
                + `${mm}`.padStart(2,'0') + ':' 
                + `${ss}`.padStart(2,'0');
    }

    drawAgentRays(newAgent) {
        let agentPosition = newAgent.pos;
        let rayLengths = newAgent.raysValues;
        if (!this.rayLines) {
            this.rayLines = new PIXI.Graphics();
            this.particleContainer.addChild(this.rayLines);
        }
    
        this.rayLines.clear();

        let sameTypeColor;
        let otherColors;
        let maxRayLength;

        let rayAngleStep;
        let currentAngle;

        let angleRadian
        let startAngle
        let angleDelta

        if (newAgent.color === "Red") {
            sameTypeColor = 0xFF0000;
            otherColors = 0x00FF00;
            maxRayLength = this.PredatorRayMaxLength

            angleRadian = this.predatorRayAngleDeg * (Math.PI / 180);
            startAngle = Math.atan2(newAgent.vel[1], newAgent.vel[0]) - angleRadian / 2;
            angleDelta = angleRadian / rayLengths.length;

            rayAngleStep = this.predatorRayAngleDeg / rayLengths.length;
        } else {
            sameTypeColor = 0x00FF00;
            otherColors = 0xFF0000;
            maxRayLength = this.PreyRayMaxLength

            angleRadian = this.preyRayAngleDeg * (Math.PI / 180);
            startAngle = Math.atan2(newAgent.vel[1], newAgent.vel[0]) - angleRadian / 2;
            angleDelta = angleRadian / (rayLengths.length - 1);

            rayAngleStep = this.preyRayAngleDeg / rayLengths.length;
        }

        let counter = 0;
        rayLengths.forEach((length) => {
            length = Number(length);
            let lineColor;
            if (length > 1000){
                lineColor = sameTypeColor;
                length -= 1000;
            } else if (length > 0){
                lineColor = otherColors;
            } else if (length === 0){
                lineColor = 0xFFFFFF; // Default color or any other color
                length = maxRayLength * SCALING;
            }

            // Set the line style with the appropriate color before drawing each line
            this.rayLines.lineStyle(1, lineColor, 1);

            let baseVector = [1,0]

            let angle = startAngle + angleDelta * counter;

            let endX = baseVector[0]*Math.cos(angle) - baseVector[1]*Math.sin(angle)
            let endY = baseVector[0]*Math.sin(angle) + baseVector[1]*Math.cos(angle)

            let endPosX = endX * length + agentPosition[0]
            let endPosY = endY * length + agentPosition[1]

            this.rayLines.moveTo(agentPosition[0], agentPosition[1]);
            this.rayLines.lineTo(endPosX, endPosY);

            currentAngle += rayAngleStep;
            counter++;
        });
    }

    drawDirectionLine(newAgent) {
        let agentPosition = newAgent.pos;
        let velocity = newAgent.vel;
        if (!this.directionLine) {
            this.directionLine = new PIXI.Graphics();
            this.particleContainer.addChild(this.directionLine);
        }
    
        this.directionLine.clear();
        this.directionLine.lineStyle(2, 0x0000FF, 1);
        this.directionLine.moveTo(agentPosition[0], agentPosition[1]);
        const endPosX = agentPosition[0] + velocity[0] * 20;
        const endPosY = agentPosition[1] + velocity[1] * 20;
        this.directionLine.lineTo(endPosX, endPosY);
    }

    drawNeuralNetwork(newAgent) {
        this.neuralNetworkContainer.x = window.innerWidth * 0.9;
        this.neuralNetworkContainer.y = window.innerHeight / 1.7;
        this.neuralNetworkContainer.removeChildren();

        // ASUME NEURONS ARE SORTED BY DEPTH
        const neurons = newAgent.brain.neurons;
        const connections = newAgent.brain.connections;

        const xSpacing = 40;
        const ySpacing = 6;

        let depth = 0;
        let currentX = 0;
        let currentY = 0;
        let mul = 1;

        const neuronPositions = new Map();
        neurons.forEach(neuron => {
            if (neuron.depth > depth) {
                depth++;
                currentX += xSpacing;
                currentY = 0;
            }
            neuronPositions.set(neuron.id, {
                x: currentX,
                y: currentY * mul
            });
            currentY += ySpacing;
            mul = -mul;
        });

        connections.forEach(connection => {
            const sourcePos = neuronPositions.get(connection.source);
            const targetPos = neuronPositions.get(connection.target);
        
            const line = new PIXI.Graphics();
        
            let normalizedWeight = connection.weight / 10;
            normalizedWeight = Math.max(-1, Math.min(normalizedWeight, 1));

            let color;
            if (normalizedWeight > 0) {
                color = 0x00FF00 + (Math.floor(255 * (1 - normalizedWeight)) << 16);
            } else {
                color = 0xFF0000 + (Math.floor(255 * (1 + normalizedWeight)) << 8);
            }
        
            line.lineStyle(2, color);
            line.moveTo(sourcePos.x, sourcePos.y);
            line.lineTo(targetPos.x, targetPos.y);
            this.neuralNetworkContainer.addChild(line);
        });

        // Draw neurons
        neurons.forEach(neuron => {
            const position = neuronPositions.get(neuron.id);
            const circle = new PIXI.Graphics();

            if (neuron.value > 100) {
                neuron.value = 0;
            } else if (neuron.value < -100) {
                neuron.value = -100;
            }

            const normalizedValue = (neuron.value + 1) / 2;

            let red, green, blue;

            if (neuron.value < 0) {
                blue = 255;
                red = green = Math.floor(normalizedValue * 2 * 255);
            } else {
                red = 255;
                green = blue = 255 - Math.floor((normalizedValue - 0.5) * 2 * 255);
            }

            const colorHex = (red << 16) + (green << 8) + blue;

            circle.beginFill(colorHex);

            circle.drawCircle(position.x, position.y, 6);
            circle.endFill();

            this.neuralNetworkContainer.addChild(circle);

            if (depth > 3) {
                this.neuralNetworkContainer.x = window.innerWidth  - currentX - 30;
                //this.neuralNetworkContainer.y = window.innerHeight / 2;
            }
            
        });
    }

}

document.addEventListener("DOMContentLoaded", function() {
    // get config from server
    const config = fetch('http://localhost:8080/config', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }}).then(response => response.json().then(data => {
            console.log(data);  
            const app = new Application(data.numAgents, data.height * SCALING, data.width * SCALING, data.cellSize * SCALING, data.agentRadius * SCALING * 2, data.predatorRayAngleDeg, data.preyRayAngleDeg,data.predatorRayLength, data.preyRayLength);
            app.initialize();
        })
    );
});
