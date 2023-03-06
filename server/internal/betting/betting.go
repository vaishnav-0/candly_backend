package betting

type prediction int
const poolPrefix string = "pool:"

const (
	RedPredict prediction = 0
	GreenPredict prediction = 1
)

type BetData struct{
	
}


func CreatePool(id string){

}


func Bet(id string, user string, amount int, pred prediction){

}

func GetBets(id string){

}