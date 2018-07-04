from flask import Flask, render_template, request, make_response, url_for, redirect
import requests, json

app = Flask(__name__)
app.config['DEBUG'] = True

@app.route('/')
def index():
    return render_template('test_page.html')

@app.route('/result', methods = ['POST', 'GET'])
def result():
    if request.method == 'POST':
        
        result = request.form
        t_lang = result['lang']
        t_code = result['code']
        url = "http://localhost:8080/api/execute"
        raw_data = {'Language': str(t_lang), 'Input': str(t_code)+ '\n'}
        print(json.dumps(raw_data))
        response = requests.post(
            url, data=json.dumps(raw_data),
            headers={'Content-Type': 'application/json'}
            )
        response.encoding = 'utf-8'
        print(response.status_code)
        print(type(response.json()))
        return render_template('answer.html', result = json.loads(response.json()))
       
        
        

app.run(host='0.0.0.0', port= 8090)
